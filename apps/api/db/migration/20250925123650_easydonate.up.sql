CREATE TABLE IF NOT EXISTS ed_products (
    product_id uuid NOT NULL,
    ed_product_id bigint NOT NULL,
    PRIMARY KEY (product_id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ed_products_ed_product_id ON ed_products(ed_product_id);

ALTER TABLE orders ADD COLUMN external_id tinytext;

-- Upgrade: Season Access
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('4086268a-34fe-47ec-afab-a014a2f02688', 1018684);
-- Name Color: Memariani
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('21852047-a4e9-4f7d-acfd-9fce898bc03d', 1018723);
-- Name Color: PacificDream
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('6509dcd7-6177-4e39-8474-e0304870ff23', 1018724);
-- Name Color: SteelGray
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('f1740c10-850a-41a6-bcd8-d3449846f284', 1018725);
-- Name Color: Dracula
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('1cef33ed-08e5-49c5-9858-dac85872386e', 1018726);
-- Name Color: Kyoto
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('6a213c8c-4269-4329-9875-917cb3cfd033', 1018727);
-- Name Color: WheatField
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('068e6c5a-94e1-4927-9eed-f9b02beb9f64', 1018728);
-- Name Color: RedSquare
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('8bbf4210-24fa-4382-af93-4da30dfeae1f', 1018729);
-- Name Color: RedMist
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('54075756-222b-4917-9924-696436654254', 1018730);
-- Name Color: MiceElf
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('72150255-2649-4339-895b-44142701117d', 1018731);
-- Name Color: Timber
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('f555a7d6-c52f-4f22-8183-cdfee77eb59e', 1018732);
-- Name Color: Lawrencium
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('56ef1088-6ce5-4cb6-9c3d-0c35070def83', 1018733);
-- Name Color: Magic
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('8f461be1-9320-4d17-96e7-e5215fc52522', 1018734);
-- Name Color: PinkFlavour
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('383c4fc1-5f92-4bae-a4b4-b743922f1eba', 1018735);
-- Name Color: CocoaaIce
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('bdd07172-ea5f-4a0a-b82a-16776a6b4b5c', 1018736);
-- Name Color: Dusk
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('9765c31a-c5d9-4d0a-9eb7-3c10aec6e66c', 1018737);
-- Name Color: KyeMeh
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('68a560a1-878b-4cee-b79d-465aea06862b', 1018738);
-- Name Color: VelvetSun
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('69c9ef9e-f66a-4df7-9b3b-bdb2dbf97834', 1018739);
-- Name Color: Telegram
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('66678f14-e431-41d1-a172-d301564deba0', 1018740);
-- Name Color: SiriusTamed
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('3189113f-9253-4bef-9ce4-0374aff2df7c', 1018741);
-- Name Color: Moonrise
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('de397bbd-782c-4eb2-a59c-f955cc06340e', 1018742);
-- Name Color: Frozen
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('63406c8d-67c2-43c1-a8a7-e7e79b7ac2c3', 1018743);
-- Name Color: DeepSpace
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('d9607e73-ad81-46e6-b0af-ee33b5987e16', 1018744);
-- Name Color: Mango
INSERT INTO ed_products (product_id, ed_product_id) VALUES ('2a65d174-8896-411a-a6dd-8341b3620171', 1018745);
