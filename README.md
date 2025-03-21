# 🛡️ gRPC Auth Service

A lightweight and secure authentication microservice built with **Go**, **gRPC**, and **SQLite**.  
Handles user registration and login with hashed passwords and a token-based response system.

---

## ✨ Features

- ✅ User registration with bcrypt password hashing
- 🔐 Secure login with JWT generation
- ⚡ High-performance gRPC API
- 🗃️ SQLite for lightweight local storage

---

## 🧰 Tech Stack

- **Go 1.24**
- **gRPC**
- **Protocol Buffers**
- **SQLite**
- **bcrypt**
- **JWT**

  ## 📦 API Definition

```proto

service Auth {
    rpc Register (RegisterRequest) returns (RegisterResponse);
    rpc LogIn (LoginRequest) returns (LoginResponse);
    rpc IsAdmin (IsAdminRequest) returns (IsAdminResponse);
}

message RegisterRequest {
    string email = 1;
    string password = 2;
}

message RegisterResponse {
    int64 user_id = 1;
}

message LoginRequest {
    string email = 1;
    string password = 2;
    int32 app_id = 3;
}

message LoginResponse {
    string token = 1;
}

message IsAdminRequest {
    int64 user_id = 1;
}

message IsAdminResponse {
    bool is_admin = 1;
}
