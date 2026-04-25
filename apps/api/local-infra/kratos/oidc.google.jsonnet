local claims = {
  email_verified: true,
} + std.extVar('claims');

{
  identity: {
    traits: {
      [if 'email' in claims && claims.email_verified then 'email' else null]: claims.email,
      username: if 'preferred_username' in claims then claims.preferred_username else '',
      avatarUrl: if 'picture' in claims then claims.picture else ''

      // Google can return the user's full name as a single string.
      // Make sure to adjust your identity schema to store the first and last name as separate traits.
      // first_name: claims.given_name,
      // last_name: claims.family_name,
      //
      // hd is the hosted domain of the user's email address. 
      // [if 'hd' in claims && claims.email_verified then 'hd' else null]: claims.hd,
    },
  },
}