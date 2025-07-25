# API Documentation - JWT Authentication System

## Base URL
```
http://localhost:8080/api/v1
```

## Overview
This API provides JWT-based authentication with email verification, refresh tokens, and user management. All endpoints return JSON responses.

## Authentication Flow
1. **Register** → User account created with `is_verified = false`
2. **Email Verification** → User clicks verification link to activate account
3. **Login** → Returns access token (15 min) + refresh token (7 days)
4. **Access Protected Routes** → Use Bearer token in Authorization header
5. **Refresh Token** → Get new access token when expired
6. **Logout** → Blacklist current token

---

## 🔐 Authentication Endpoints

### 1. Register User
**POST** `/auth/register`

Create a new user account. User will be registered with `is_verified = false` and must verify email before login.

#### Request Body
```json
{
  "email": "user@example.com",
  "password": "securepassword123",
  "first_name": "John",
  "last_name": "Doe"
}
```

#### Response (201 Created)
```json
{
  "message": "User registered successfully. Please check your email to verify your account."
}
```

#### Response (400 Bad Request)
```json
{
  "error": "user with this email already exists"
}
```

---

### 2. Verify Email
**GET** `/auth/verify-email?token={verification_token}`

Verify user's email address using the token sent via email.

#### Query Parameters
- `token` (required): Email verification token from the verification email

#### Response (200 OK)
```json
{
  "message": "Email verified successfully. You can now log in."
}
```

#### Response (400 Bad Request)
```json
{
  "error": "invalid or expired verification token"
}
```

---

### 3. Resend Verification Email
**POST** `/auth/resend-verification`

Resend verification email to user if they didn't receive it or token expired.

#### Request Body
```json
{
  "email": "user@example.com"
}
```

#### Response (200 OK)
```json
{
  "message": "Verification email sent successfully."
}
```

#### Response (400 Bad Request)
```json
{
  "error": "email already verified"
}
```

---

### 4. User Login
**POST** `/auth/login`

Authenticate user and receive JWT tokens. User must have verified email.

#### Request Body
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

#### Response (200 OK)
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0",
  "expires_in": 900,
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "is_verified": true,
    "is_active": true,
    "role_id": 2,
    "created_at": "2025-07-26T00:49:19Z",
    "updated_at": "2025-07-26T00:49:19Z"
  }
}
```

#### Response (401 Unauthorized)
```json
{
  "error": "invalid email or password"
}
```

#### Response (403 Forbidden)
```json
{
  "error": "please verify your email address before logging in"
}
```

---

### 5. Refresh Access Token
**POST** `/auth/refresh-token`

Get a new access token using a valid refresh token.

#### Request Body
```json
{
  "refresh_token": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0"
}
```

#### Response (200 OK)
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "z9y8x7w6v5u4t3s2r1q0p9o8n7m6l5k4j3i2h1g0",
  "expires_in": 900,
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "is_verified": true,
    "is_active": true,
    "role_id": 2,
    "created_at": "2025-07-26T00:49:19Z",
    "updated_at": "2025-07-26T00:49:19Z"
  }
}
```

#### Response (401 Unauthorized)
```json
{
  "error": "refresh token expired"
}
```

---

### 6. Logout
**POST** `/auth/logout`

Logout user and blacklist current access token.

#### Headers
```
Authorization: Bearer {access_token}
```

#### Response (200 OK)
```json
{
  "message": "Logout successful"
}
```

#### Response (401 Unauthorized)
```json
{
  "error": "invalid or expired token"
}
```

---

### 7. Get User Profile
**GET** `/auth/profile`

Get current authenticated user's profile information.

#### Headers
```
Authorization: Bearer {access_token}
```

#### Response (200 OK)
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "is_verified": true,
  "is_active": true,
  "role_id": 2,
  "created_at": "2025-07-26T00:49:19Z",
  "updated_at": "2025-07-26T00:49:19Z"
}
```

#### Response (401 Unauthorized)
```json
{
  "error": "invalid or expired token"
}
```

---

## 🛡️ Protected Routes

All protected routes require the `Authorization` header with a valid JWT token:

```
Authorization: Bearer {access_token}
```

### User Management Endpoints

#### Get All Users
**GET** `/users/`

#### Get User by ID
**GET** `/users/{id}`

#### Update User
**PUT** `/users/{id}`

#### Delete User
**DELETE** `/users/{id}`

---

## 🔧 System Endpoints

### Health Check
**GET** `/health`

Check if the API is running.

#### Response (200 OK)
```json
{
  "status": "OK",
  "message": "API is running"
}
```

---

## 📊 Data Models

### User Model
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "is_verified": true,
  "is_active": true,
  "role_id": 2,
  "created_at": "2025-07-26T00:49:19Z",
  "updated_at": "2025-07-26T00:49:19Z"
}
```

### Auth Response Model
```json
{
  "access_token": "JWT_TOKEN_HERE",
  "refresh_token": "REFRESH_TOKEN_HERE",
  "expires_in": 900,
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "is_verified": true,
    "is_active": true,
    "role_id": 2,
    "created_at": "2025-07-26T00:49:19Z",
    "updated_at": "2025-07-26T00:49:19Z"
  }
}
```

---

## ⚙️ Configuration

### Token Expiration
- **Access Token**: 15 minutes
- **Refresh Token**: 7 days
- **Email Verification Token**: 24 hours

### JWT Claims
```json
{
  "sub": 1,           // User ID
  "exp": 1721952559,  // Expiration timestamp
  "iat": 1721951659,  // Issued at timestamp
  "jti": "random-id", // JWT ID for blacklisting
  "type": "access"    // Token type
}
```

---

## 🚨 Error Responses

All error responses follow this format:

```json
{
  "error": "Error message describing what went wrong"
}
```

### Common HTTP Status Codes
- `200` - Success
- `201` - Created (successful registration)
- `400` - Bad Request (validation errors, missing fields)
- `401` - Unauthorized (invalid credentials, expired token)
- `403` - Forbidden (unverified email, insufficient permissions)
- `404` - Not Found (user or resource not found)
- `500` - Internal Server Error

---

## 📧 Email Verification Flow

1. **User registers** → Account created with `is_verified = false`
2. **Verification email sent** → Contains secure token (32-byte hex)
3. **User clicks link** → `GET /auth/verify-email?token={token}`
4. **Token validated** → User's `is_verified` set to `true`
5. **User can login** → Email verification required for login

### Email Development Mode
In development, emails are logged to console instead of being sent via SMTP.

---

## 🔐 Security Features

- **Password Hashing**: bcrypt with salt
- **Secure Token Generation**: crypto/rand with 32-byte tokens
- **JWT Signing**: HMAC SHA-256
- **Token Blacklisting**: Prevents token reuse after logout
- **Token Rotation**: New refresh token issued on each refresh
- **Email Verification**: Required before login
- **Request Logging**: All auth attempts logged with IP/User-Agent

---

## 📝 Frontend Integration Examples

### JavaScript/Fetch Example

```javascript
// Register User
const register = async (userData) => {
  const response = await fetch('/api/v1/auth/register', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(userData)
  });
  return response.json();
};

// Login User
const login = async (credentials) => {
  const response = await fetch('/api/v1/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(credentials)
  });
  const data = await response.json();
  
  if (response.ok) {
    localStorage.setItem('access_token', data.access_token);
    localStorage.setItem('refresh_token', data.refresh_token);
  }
  
  return data;
};

// Make Authenticated Request
const makeAuthenticatedRequest = async (url, options = {}) => {
  const token = localStorage.getItem('access_token');
  
  const response = await fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    }
  });
  
  if (response.status === 401) {
    // Token expired, try to refresh
    const refreshed = await refreshToken();
    if (refreshed) {
      // Retry original request with new token
      return makeAuthenticatedRequest(url, options);
    } else {
      // Refresh failed, redirect to login
      window.location.href = '/login';
    }
  }
  
  return response.json();
};

// Refresh Token
const refreshToken = async () => {
  const refresh_token = localStorage.getItem('refresh_token');
  
  const response = await fetch('/api/v1/auth/refresh-token', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ refresh_token })
  });
  
  if (response.ok) {
    const data = await response.json();
    localStorage.setItem('access_token', data.access_token);
    localStorage.setItem('refresh_token', data.refresh_token);
    return true;
  } else {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    return false;
  }
};
```

---

## 🐛 Troubleshooting

### Common Issues

1. **"please verify your email address before logging in"**
   - User hasn't clicked verification link
   - Use resend verification endpoint

2. **"invalid or expired token"**
   - Access token expired (15 minutes)
   - Use refresh token to get new access token

3. **"refresh token expired"**
   - Refresh token expired (7 days)
   - User must login again

4. **"user with this email already exists"**
   - Email already registered
   - User should login or reset password

5. **"invalid or expired verification token"**
   - Verification token expired (24 hours)
   - Use resend verification endpoint

---

## 📞 Support

For any questions or issues with the API, please contact the backend development team.

**Server Status**: ✅ Running on http://localhost:8080
**Database**: ✅ MySQL Connected
**Email Service**: ✅ Console Logging (Development Mode)
