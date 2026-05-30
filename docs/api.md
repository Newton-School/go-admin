# JSON API

The API is versioned under `/admin/api/v1`.

Routes:

- `GET /admin/api/v1/apps`
- `GET /admin/api/v1/{app}/{resource}`
- `POST /admin/api/v1/{app}/{resource}`
- `GET /admin/api/v1/{app}/{resource}/{id}`
- `PATCH /admin/api/v1/{app}/{resource}/{id}`
- `PUT /admin/api/v1/{app}/{resource}/{id}`
- `DELETE /admin/api/v1/{app}/{resource}/{id}`
- `POST /admin/api/v1/{app}/{resource}/actions/{action}`
- `GET /admin/api/v1/{app}/{resource}/lookup/{field}?q=term`

The SDK does not authenticate API calls. Protect the mounted handler with host middleware.

