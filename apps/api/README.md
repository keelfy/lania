# Lania API

## Admin panel access

Admin routes live under `/v1/admin`. They require an active Ory session whose identity has
`metadata_public.role` set to `owner` or `admin`. Any other identity gets `403`.

The API reads and writes identities through the Kratos admin API. Set `ORY_ADMIN_URL` to the
admin address (`http://kratos:4434` in the compose stack). Never expose that address outside the
private network.

### Make the first admin

The role can only be written with the Kratos admin API. Run this on the server, which puts the
request inside the Docker network:

```bash
docker run --rm --network lania-web-net curlimages/curl \
  -X PATCH http://kratos:4434/admin/identities/<identity-id> \
  -H 'Content-Type: application/json' \
  -d '[{"op":"add","path":"/metadata_public","value":{"role":"owner"}}]'
```

The identity ID is the `owner_user_id` of any profile of that user. This replaces
`metadata_public`, so include every key the identity already has in it.
