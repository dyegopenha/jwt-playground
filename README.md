# JWT Playground

This project is a simple JWT playground in Go. It provides a simple API for signing in, refreshing tokens, and accessing a protected profile endpoint.

## 🚀 Running the project

To run the project, you need to have [Go](https://go.dev/doc/install) installed.

Run the server with the following command:

```bash
make run
```

This will start the server on the port specified in the `.env` file (defaults to `8080`).

## 📝 API Endpoints

The following endpoints are available:

### Public Endpoints

#### `POST /sign-in`

This endpoint allows you to sign in and get a JWT.

**Request body:**

```json
{
  "email": "user@email.com",
  "password": "password"
}
```

**Response:**

```json
{
  "access_token": "..."
}
```

#### `POST /refresh`

This endpoint allows you to refresh your JWT.

**Response:**

```json
{
  "access_token": "..."
}
```

### Protected Endpoints

#### `GET /`

This endpoint returns the user's profile. You need to provide a valid JWT in the `Authorization` header.

**Request header:**

```
Authorization: Bearer <access_token>
```

**Response:**

```json
{
  "id": "...",
  "role": "..."
}
```
