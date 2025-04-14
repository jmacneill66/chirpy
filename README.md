# Chirpy - A Simple Social Microblogging Platform 🐦

Chirpy is a lightweight microblogging API that allows users to create, read, and delete short messages called "chirps." The backend is built using Go and Gorilla/Mux, with SQLC for database access.
🚀 Features

    User authentication & JWT-based session management.

    Create, fetch, and delete chirps.

    Chirpy Red Membership: Paid users can edit their chirps.

    Secure webhook integration with Polka (payment provider).

    Admin controls to reset data and view metrics.

🛠 Installation & Setup
Prerequisites

    Go 1.20+ installed

    PostgreSQL installed and running

    SQLC installed

1️⃣ Clone the Repository

git clone <https://github.com/your-username/chirpy.git>
cd chirpy

2️⃣ Set Up Environment Variables

Create a .env file in the root directory:

DATABASE_URL=postgres://user:password@localhost:5432/chirpy
JWT_SECRET=your_jwt_secret
POLKA_KEY=f271c81ff7084ee5b99a5091b42d486e

3️⃣ Run Database Migrations

make migrate-up

or

go run scripts/migrate.go

4️⃣ Start the Server

go run main.go

Your Chirpy API should now be running at <http://localhost:8080> 🎉
🔥 API Endpoints
Authentication
🔹 Register User

POST /api/users

Request Body:

{
  "username": "johndoe",
  "password": "securepassword"
}

Response:

{
  "id": "3311741c-680c-4546-99f3-fc9efac2036c",
  "username": "johndoe",
  "is_chirpy_red": false
}

🔹 Login

POST /api/login

Request Body:

{
  "username": "johndoe",
  "password": "securepassword"
}

Response:

{
  "token": "your.jwt.token.here"
}

Chirps
🔹 Create a Chirp

POST /api/chirps

Request Body:

{
  "body": "Hello, Chirpy!"
}

Response:

{
  "id": "b5c124d9-7f3a-4c60-9c1e-5f9c41e5dc24",
  "user_id": "3311741c-680c-4546-99f3-fc9efac2036c",
  "body": "Hello, Chirpy!",
  "created_at": "2024-03-23T12:00:00Z"
}

🔹 Get Chirps

GET /api/chirps?sort=asc&author_id=3311741c-680c-4546-99f3-fc9efac2036c

Query Parameters:

    sort=asc|desc (default: asc)

    author_id=<user_id> (optional)

Response:

[
  {
    "id": "b5c124d9-7f3a-4c60-9c1e-5f9c41e5dc24",
    "user_id": "3311741c-680c-4546-99f3-fc9efac2036c",
    "body": "Hello, Chirpy!",
    "created_at": "2024-03-23T12:00:00Z"
  }
]

🔹 Delete a Chirp

DELETE /api/chirps/{chirpID}
Authorization: Bearer YOUR_JWT_TOKEN

Response:
204 No Content
Chirpy Red Membership
🔹 Webhook from Polka

POST /api/polka/webhooks
Authorization: ApiKey f271c81ff7084ee5b99a5091b42d486e

Request Body:

{
  "event": "user.upgraded",
  "data": {
    "user_id": "3311741c-680c-4546-99f3-fc9efac2036c"
  }
}

Response:
204 No Content (if successful)
401 Unauthorized (if API key is missing or incorrect)
⚙️ Project Structure

chirpy/
│── api/                # HTTP handlers
│── auth/               # Authentication (JWT, API keys)
│── database/           # SQLC-generated queries
│── models/             # Data models
│── scripts/            # Database migrations
│── .env.example        # Environment variables example
│── main.go             # Main server entry point
│── router.go           # Gorilla/Mux routing setup
│── README.md           # This file!

✅ Testing

Run unit tests with:

go test ./...

🤝 Contributing

    Fork the repo

    Create a new branch (git checkout -b feature-name)

    Commit your changes (git commit -m "Add new feature")

    Push to the branch (git push origin feature-name)

    Open a Pull Request

📜 License

This project is licensed under the MIT License. See LICENSE for details.
📧 Contact

For questions or issues, feel free to reach out:
✉️ Email: <jeffrey.macneill@gmail.com>


