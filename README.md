# bhagabad_gita backend.


![Screenshot from 2024-12-18 19-48-13](https://github.com/user-attachments/assets/e61ee059-5e42-46f5-926c-06f4367a52f7)


First, clone the repository.

```bash
git clone https://github.com/asifrahaman13/bhagavad_gita.git
```

## Backend

Rename the .env.example file to .env file.

In Unix based system, you can use the following:

```bash
mv .env.example .env
```

In windows you can manually do the same.

Next update the required values in the .env file.

Now run the application in local development server.

```bash
go build && go run main.go
```

The backend will run on the following port:

http://localhost:8000

## Frontend

Go to the frontend folder.

```bash
cd frontend/
```

Next install the necessary packages.

```bash
bun install
```

Now run the frontend server.

```bash
bun run dev
```

## Ports

The application will run on the following ports.

- `Backend`: http://127.0.0.1/8000
- `Frontend`: http://127.0.0.1/3000
