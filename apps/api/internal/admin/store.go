package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	storesql "github.com/lania-smp/backend/internal/storage/main"
)

type Store struct {
	DB                         *sql.DB
	ActiveSeason, DefaultColor uuid.UUID
}
type Profile struct {
	ID            string  `json:"id"`
	MinecraftUUID string  `json:"minecraftUUID"`
	Username      string  `json:"username"`
	Owner         *string `json:"ownerId"`
}

func (p *Profile) OwnerID() string {
	if p.Owner == nil {
		return ""
	}
	return *p.Owner
}

type Product struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Category domain.ProductCategory `json:"category"`
	Metadata json.RawMessage        `json:"metadata"`
}
type Season struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
}
type Grant struct {
	ID        string  `json:"id"`
	Kind      string  `json:"kind"`
	Name      string  `json:"name"`
	Season    *string `json:"season"`
	Protected bool    `json:"protected"`
}

func (s *Store) Profiles(ctx context.Context, search, owner string, offset int) ([]Profile, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id, mc_uuid, mc_username, owner_user_id FROM profiles
 WHERE (? = '' OR mc_username LIKE ? OR id = ?) AND (? = '' OR owner_user_id = ?)
 ORDER BY mc_username, id LIMIT 51 OFFSET ?`, search, "%"+search+"%", search, owner, owner, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Profile{}
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.ID, &p.MinecraftUUID, &p.Username, &p.Owner); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (s *Store) Profile(ctx context.Context, id uuid.UUID) (*Profile, error) {
	var p Profile
	err := s.DB.QueryRowContext(ctx, `SELECT id, mc_uuid, mc_username, owner_user_id FROM profiles WHERE id = ?`, id).Scan(&p.ID, &p.MinecraftUUID, &p.Username, &p.Owner)
	return &p, err
}

func (s *Store) Products(ctx context.Context) ([]Product, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT p.id, COALESCE(l.name, p.price_name), p.category, p.metadata FROM products p
 LEFT JOIN product_localizations l ON l.product_id = p.id AND l.locale = 'ru' ORDER BY p.category, p.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Product{}
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Metadata); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (s *Store) Seasons(ctx context.Context) ([]Season, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id, season_number FROM seasons ORDER BY season_number DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Season{}
	for rows.Next() {
		var season Season
		if err := rows.Scan(&season.ID, &season.Number); err != nil {
			return nil, err
		}
		result = append(result, season)
	}
	return result, rows.Err()
}

func (s *Store) Grants(ctx context.Context, p *Profile) ([]Grant, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT a.id, 'access', 'Доступ к сезону', CAST(sn.season_number AS CHAR), 0
 FROM profile_accesses a JOIN seasons sn ON sn.id = a.season_id WHERE a.mc_uuid = ?
 UNION ALL SELECT o.id, 'color', c.name, CAST(sn.season_number AS CHAR), o.name_color_id = ?
 FROM profile_name_color_options o JOIN name_colors c ON c.id = o.name_color_id LEFT JOIN seasons sn ON sn.id = o.for_season_id WHERE o.profile_id = ?
 UNION ALL SELECT o.id, 'prefix', n.name, CAST(sn.season_number AS CHAR), 0
 FROM profile_name_prefix_options o JOIN name_prefixes n ON n.id = o.name_prefix_id LEFT JOIN seasons sn ON sn.id = o.for_season_id
 WHERE o.profile_id = ? AND o.type = 'glyth'`, p.MinecraftUUID, s.DefaultColor, p.ID, p.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Grant{}
	for rows.Next() {
		var g Grant
		if err := rows.Scan(&g.ID, &g.Kind, &g.Name, &g.Season, &g.Protected); err != nil {
			return nil, err
		}
		result = append(result, g)
	}
	return result, rows.Err()
}

// Every mutation locks the profile, so concurrent admin actions cannot overwrite a stale owner.
func (s *Store) mutate(ctx context.Context, id uuid.UUID, fn func(*sql.Tx, *Profile) error) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var p Profile
	err = tx.QueryRowContext(ctx, `SELECT id, mc_uuid, mc_username, owner_user_id FROM profiles WHERE id = ? FOR UPDATE`, id).Scan(&p.ID, &p.MinecraftUUID, &p.Username, &p.Owner)
	if err != nil {
		return err
	}
	if err = fn(tx, &p); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) Owner(ctx context.Context, id, actor uuid.UUID, expected, owner string) error {
	return s.mutate(ctx, id, func(tx *sql.Tx, p *Profile) error {
		if p.OwnerID() != expected {
			return &userError{409, "Владелец изменился. Обновите страницу."}
		}
		if p.OwnerID() != "" && owner != "" && p.OwnerID() != owner {
			return &userError{409, "Сначала отвяжите профиль от текущего аккаунта."}
		}
		var value any
		if owner != "" {
			value = owner
		}
		_, err := tx.ExecContext(ctx, `UPDATE profiles SET owner_user_id = ?, updated_at = now(), updated_by = ? WHERE id = ?`, value, actor, id)
		return err
	})
}

func (s *Store) Give(ctx context.Context, id, productID, seasonID, actor uuid.UUID) error {
	return s.mutate(ctx, id, func(tx *sql.Tx, p *Profile) error {
		var season string
		if err := tx.QueryRowContext(ctx, `SELECT id FROM seasons WHERE id = ?`, seasonID).Scan(&season); err != nil {
			return err
		}
		var product Product
		if err := tx.QueryRowContext(ctx, `SELECT category, metadata FROM products WHERE id = ?`, productID).Scan(&product.Category, &product.Metadata); err != nil {
			return err
		}
		q := storesql.WithMySQLTx(tx)
		switch product.Category {
		case domain.ProductCategoryUpgrade:
			var metadata domain.UpgradeProductMetadata
			if err := json.Unmarshal(product.Metadata, &metadata); err != nil {
				return err
			}
			if metadata.Action != domain.ProductUpgradeActionSeasonAccess {
				return &userError{400, "Этот продукт не поддерживается."}
			}
			_, err := tx.ExecContext(ctx, `INSERT INTO profile_accesses (mc_uuid, season_id, source, updated_by)
    SELECT ?, ?, 'admin', ? WHERE NOT EXISTS (SELECT 1 FROM profile_accesses WHERE mc_uuid = ? AND season_id = ?)`, p.MinecraftUUID, seasonID, actor, p.MinecraftUUID, seasonID)
			return err
		case domain.ProductCategoryNameColor:
			var metadata domain.NameColorProductMetadata
			if err := json.Unmarshal(product.Metadata, &metadata); err != nil {
				return err
			}
			if metadata.NameColorID == uuid.Nil {
				return errors.New("missing product nameColorId")
			}
			var exists bool
			err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM profile_name_color_options WHERE profile_id = ? AND name_color_id = ? AND (for_season_id = ? OR for_season_id IS NULL))`, id, metadata.NameColorID, seasonID).Scan(&exists)
			if err != nil || exists {
				return err
			}
			return q.InsertProfileNameColorOption(ctx, storesql.InsertProfileNameColorOptionParams{ProfileID: id, NameColorID: metadata.NameColorID, ForSeasonID: &seasonID})
		case domain.ProductCategoryNamePrefix:
			var metadata domain.NamePrefixProductMetadata
			if err := json.Unmarshal(product.Metadata, &metadata); err != nil {
				return err
			}
			if metadata.NamePrefixID == uuid.Nil {
				return errors.New("missing product namePrefixId")
			}
			var exists bool
			err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM profile_name_prefix_options WHERE profile_id = ? AND name_prefix_id = ? AND (for_season_id = ? OR for_season_id IS NULL))`, id, metadata.NamePrefixID, seasonID).Scan(&exists)
			if err != nil || exists {
				return err
			}
			return q.InsertProfileNamePrefixOption(ctx, storesql.InsertProfileNamePrefixOptionParams{ProfileID: id, NamePrefixID: metadata.NamePrefixID, ForSeasonID: &seasonID, Type: domain.ProfilePrefixTypeGlyth})
		default:
			return &userError{400, "Этот продукт не поддерживается."}
		}
	})
}

func (s *Store) Revoke(ctx context.Context, id, grantID, actor uuid.UUID, kind string) error {
	return s.mutate(ctx, id, func(tx *sql.Tx, p *Profile) error {
		switch kind {
		case "access":
			var season string
			err := tx.QueryRowContext(ctx, `SELECT season_id FROM profile_accesses WHERE id = ? AND mc_uuid = ?`, grantID, p.MinecraftUUID).Scan(&season)
			if err != nil {
				return err
			}
			// Legacy data may contain multiple grants for one season; revocation removes access completely.
			_, err = tx.ExecContext(ctx, `DELETE FROM profile_accesses WHERE mc_uuid = ? AND season_id = ?`, p.MinecraftUUID, season)
			return err
		case "color", "prefix":
			table, column := "profile_name_color_options", "name_color_id"
			if kind == "prefix" {
				table, column = "profile_name_prefix_options", "name_prefix_id"
			}
			var cosmetic uuid.UUID
			query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ? AND profile_id = ?", column, table)
			if kind == "prefix" {
				query += " AND type = 'glyth'"
			}
			if err := tx.QueryRowContext(ctx, query, grantID, id).Scan(&cosmetic); err != nil {
				return err
			}
			if kind == "color" && cosmetic == s.DefaultColor {
				return &userError{400, "Базовый цвет нельзя отозвать."}
			}
			if _, err := tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE id = ? AND profile_id = ?", grantID, id); err != nil {
				return err
			}
			var remains bool
			query = fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE profile_id = ? AND %s = ? AND (for_season_id IS NULL OR for_season_id = ?))", table, column)
			if err := tx.QueryRowContext(ctx, query, id, cosmetic, s.ActiveSeason).Scan(&remains); err != nil {
				return err
			}
			if remains {
				return nil
			}
			if kind == "color" {
				_, err := tx.ExecContext(ctx, `UPDATE profiles SET name_color_id = ?, updated_at = now(), updated_by = ? WHERE id = ? AND name_color_id = ?`, s.DefaultColor, actor, id, cosmetic)
				return err
			}
			_, err := tx.ExecContext(ctx, `DELETE FROM profile_prefixes WHERE profile_id = ? AND name_prefix_id = ? AND type = 'glyth'`, id, cosmetic)
			return err
		default:
			return &userError{400, "Неизвестный тип продукта."}
		}
	})
}
