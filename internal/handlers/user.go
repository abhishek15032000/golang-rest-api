package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"rest-api/internal/dtos"
	"rest-api/internal/middleware"
	"rest-api/internal/store"
	"rest-api/internal/utils"
	"rest-api/internal/validation"
	"strconv"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/golang-jwt/jwt/v5"
	"gopkg.in/gomail.v2"
)

// upload User profile

func (h *Handler) UploadProfileImage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(middleware.UserClaimsKey).(jwt.MapClaims)
		if !ok {
			utils.RespondWithError(w, http.StatusBadRequest, "please login to continue")
			return
		}
		userID := int32(claims["user_id"].(float64))
		// upload from the form data
		err := r.ParseMultipartForm(10 << 20) // 10MB
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Failed to parse form")
			return
		}
		// file type : image/file
		file, fileHeader, err := r.FormFile("profile_image")
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Failed to get file")
			return
		}
		defer file.Close()
		// check file type
		cld, err := cloudinary.NewFromParams(os.Getenv("CLOUDINARY_CLOUD_NAME"), os.Getenv("CLOUDINARY_API_KEY"), os.Getenv("CLOUDINARY_API_SECRET"))
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to initialize Cloudinary")
			return
		}
		// save file
		uploadedResult, err := cld.Upload.Upload(r.Context(), file, uploader.UploadParams{
			PublicID: fileHeader.Filename,
			Folder:   "profile_image",
		})
		// fmt.Println(uploadedResult)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to upload file")
			return
		}
		// commit to the db
		h.Queries.CreateUserProfile(r.Context(), store.CreateUserProfileParams{
			UserID:       int32(userID),
			ProfileImage: sql.NullString{String: uploadedResult.SecureURL, Valid: uploadedResult.SecureURL != ""},
		})
		utils.RespondWithSuccess(w, http.StatusOK, "Profile image uploaded successfully", uploadedResult.SecureURL)
	}
}

// login a user.

func (h *Handler) LoginUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req dtos.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		// validate the request
		if err := validation.Validate(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		user, err := h.Queries.GetUserByUsername(ctx, req.Username)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid username or password")
			return
		}

		if !utils.ComparePassword(req.Password, user.Password) {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid username or password")
			return
		}

		token, err := utils.GenerateJWT(int(user.ID), user.Username)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
			return
		}

		// i will be creating a refresh token here.
		refreshToken, err := utils.Generaterefreshtoken(user.ID, h.Queries, &ctx)
		if err != nil {
			// fmt.Println(err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to generate refresh token")
			return
		}

		cookie := &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			HttpOnly: true,
			Secure:   false,                   // but here i am using http only. i should use secure true only when using https.
			SameSite: http.SameSiteStrictMode, // it is restricting this cookie to only be send to my server and no other server.
			Path:     "/",
			MaxAge:   60 * 60 * 24 * 14, // so 2 weeks
		}

		utils.RespondSuccessfullLogin(w, http.StatusOK, "User logged in successfully", map[string]string{
			"token": token,
		}, cookie)
	}
}

// to create a refresh token you need to login. but to refresh the access token, you need to hit this api which i am about to create with refresh token,
// so that i could provide you one access token.i will be setting this refresh token as a cookie, so that it get also send by the browser in all
// requests to the server.

// once he logs out, i will blacklist the refresh token as well, so that his access gets immediately revoked, even if he has a valid access token.

// createUser with transaction
func (h *Handler) RefreshAccessToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "No refresh token found")
			return
		}
		// now i will generate you one refresh token and send you back the token.
		// the query is written in such a way that if the token hash exist and has not expired then only i will get a record else no record.
		token, err := h.Queries.GetRefreshToken(ctx, cookie.Value)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid refresh token, login to continue")
			return
		}
		// if it valid then we will generate a new access token and send it to the user.
		// i need to get names as well from that user id.
		user_details, err := h.Queries.GetUser(ctx, token.UserID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user details")
			return
		}
		access_token, err := utils.GenerateJWT(int(token.UserID), user_details.Username)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
			return
		}
		// revoke the current refresh token.
		_, err = h.Queries.MarkRefreshTokenRevoked(ctx, cookie.Value)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to revoke refresh token")
			return
		}

		// i will be creating a refresh token here.
		refreshToken, err := utils.Generaterefreshtoken(token.UserID, h.Queries, &ctx)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to generate refresh token")
			return
		}

		new_cookie := &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			HttpOnly: true,
			Secure:   false,                   // but here i am using http only. i should use secure true only when using https.
			SameSite: http.SameSiteStrictMode, // it is restricting this cookie to only be send to my server and no other server.
			Path:     "/",
			MaxAge:   60 * 60 * 24 * 14, // so 2 weeks
		}

		// blacklist previous access token
		tokenString := extractTokenFromHander(r)
		if tokenString == "" {
			utils.RespondWithError(w, http.StatusUnauthorized, "token not found")
			return
		}

		if err := h.Redis.Set(ctx, tokenString, "blacklisted", 10*time.Minute).Err(); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to blacklist token")
			return
		}

		utils.RespondSuccessfullLogin(w, http.StatusOK, "Access token generated successfully", map[string]string{
			"token": access_token,
		}, new_cookie)
	}
}

func (h *Handler) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req dtos.CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		// Validate the request
		if err := validation.Validate(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Start a transaction
		tx, err := h.DB.BeginTx(ctx, nil)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to start transaction")
			return
		}
		defer tx.Rollback() // Ensures rollback if any error occurs

		// Create a new Queries instance bound to the transaction
		qtx := store.New(tx)

		// Check if the username already exists
		_, err = qtx.GetUserByUsername(ctx, req.Username)
		if err == nil {
			utils.RespondWithError(w, http.StatusConflict, "Username already taken")
			return
		}
		if err != sql.ErrNoRows {
			utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Database error on username check: %v", err))
			return
		}

		// Check if the email already exists
		_, err = qtx.GetUserByEmail(ctx, req.Email)
		if err == nil {
			utils.RespondWithError(w, http.StatusConflict, "Email already taken")
			return
		}
		// Check if it's a real DB error or just "Not Found"
		if err != sql.ErrNoRows {
			utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Database error on email check: %v", err))
			return
		}

		// Hash password
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		// Create user within the transaction
		_, err = qtx.CreateUser(ctx, store.CreateUserParams{
			Username: req.Username,
			Email:    req.Email,
			Password: string(hashedPassword),
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		// Commit the transaction if all operations succeed
		if err := tx.Commit(); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
			return
		}

		utils.RespondWithSuccess(w, http.StatusCreated, "User created successfully", nil)
	}
}

func (h *Handler) GetProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := r.Context().Value(middleware.UserClaimsKey).(jwt.MapClaims)
		if !ok {
			utils.RespondWithError(w, http.StatusBadRequest, "please login to continue")
			return
		}
		userID := int32(claims["user_id"].(float64))
		// check the redis first
		cacheKey := fmt.Sprintf("user:%d", userID)
		if cached, err := h.Redis.Get(ctx, cacheKey).Result(); err == nil {
			var user store.User
			if err := json.Unmarshal([]byte(cached), &user); err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to unmarshal cached data")
				return
			}
			utils.RespondWithSuccess(w, http.StatusOK, "User profile fetched successfully (success from cache/redis)", user)
			return
		}
		// fall back to db;
		user, err := h.Queries.GetUser(ctx, userID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "user not found")
			return
		}
		// set in redis.
		userJSON, _ := json.Marshal(user) // converting into json string
		// marshal means - converting the struct into json string
		// unmarshal means - converting the json to struct
		h.Redis.Set(ctx, cacheKey, userJSON, 5*time.Minute)
		utils.RespondWithSuccess(w, http.StatusOK, "User profile fetched successfully", user)
	}
}

func (h *Handler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. extract jwt claims from the request context
		ctx := r.Context()
		claims, ok := r.Context().Value(middleware.UserClaimsKey).(jwt.MapClaims)
		if !ok {
			utils.RespondWithError(w, http.StatusBadRequest, "please login to continue")
			return
		}
		// extract token from auth header
		tokenString := extractTokenFromHander(r)
		if tokenString == "" {
			utils.RespondWithError(w, http.StatusUnauthorized, "token not found")
			return
		}
		// 2. convert expiresAt to time.Time
		expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
		currentTIme := time.Now()
		ttl := expirationTime.Sub(currentTIme)
		if ttl <= 0 {
			ttl = 5 * time.Minute // fallback ttl
		}
		// 3. add token to blacklist in redis
		// it is setting tokenstring:blacklisted like this in redis
		if err := h.Redis.Set(ctx, tokenString, "blacklisted", ttl).Err(); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to blacklist token")
			return
		}
		// revoke the refresh token when logged out.
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "please login to continue")
			return
		}
		_, err = h.Queries.MarkRefreshTokenRevoked(ctx, cookie.Value)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to revoke refresh token")
			return
		}
		// clean user cache in redis:-
		cacheKey := fmt.Sprintf("user:%d", int32(claims["user_id"].(float64)))
		if err := h.Redis.Del(ctx, cacheKey).Err(); err != nil {
			fmt.Printf("Error Cleaning user cache for %s in redis: %v\n", cacheKey, err)
		}
		utils.RespondWithSuccess(w, http.StatusOK, "User logged out successfully", nil)
	}
}

func extractTokenFromHander(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

func (h *Handler) ListAllUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		users, err := h.Queries.ListUsers(ctx, store.ListUsersParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to list users")
			return
		}
		utils.RespondWithSuccess(w, http.StatusOK, "Users listed successfully", users)
	}
}

func (h *Handler) VerifyEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		claims, ok := r.Context().Value(middleware.UserClaimsKey).(jwt.MapClaims)
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "please login to continue")
			return
		}

		uid, ok := claims["user_id"].(float64)
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		userID := int32(uid)

		user, err := h.Queries.GetUser(ctx, userID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "user not found")
			return
		}

		if user.EmailVerified.Bool {
			utils.RespondWithError(w, http.StatusBadRequest, "email already verified")
			return
		}

		otp, err := utils.GenerateOTP()
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "failed to generate OTP")
			return
		}

		otp_hash, err := utils.HashOTP(otp)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "failed to generate OTP")
			return
		}

		// store OTP first
		_, err = h.Queries.CreateOTP(ctx, store.CreateOTPParams{
			OtpKey: otp_hash,
			UserID: userID,
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "failed to create OTP")
			return
		}

		from := os.Getenv("FROM_EMAIL")
		appPassword := os.Getenv("SMTP_PASSWORD")
		smtpHost := os.Getenv("SMTP_SERVER")

		htmlString, err := utils.GetEmailTemplate(otp, user.Username, "The Go Project", 5)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "failed to generate email")
			return
		}

		m := gomail.NewMessage()
		m.SetHeader("From", from)
		m.SetHeader("To", user.Email)
		m.SetHeader("Subject", "Verify Your Email")
		m.SetBody("text/html", htmlString)

		// d := gomail.NewDialer(smtpHost, 587, from, appPassword)
		// if err := d.DialAndSend(m); err != nil {
		// 	utils.RespondWithError(w, http.StatusInternalServerError, "failed to send email")
		// 	return
		// }

		go func(recipient, htmlBody string, from, appPassword, smtpHost string) {
			m := gomail.NewMessage()
			// Assume these are loaded into the Handler struct at startup
			m.SetHeader("From", from)
			m.SetHeader("To", recipient)
			m.SetHeader("Subject", "Verify Your Email")
			m.SetBody("text/html", htmlBody)
			d := gomail.NewDialer(smtpHost, 587, from, appPassword)
			if err := d.DialAndSend(m); err != nil {
				// Log the error (don't fail the HTTP request since it already returned)
				fmt.Printf("Failed to send background email to %s: %v\n", recipient, err)
			}
		}(user.Email, htmlString, from, appPassword, smtpHost)

		utils.RespondWithSuccess(w, http.StatusOK, "Email sent successfully", nil)
	}
}

func (h *Handler) VerifyOTP() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// i will do this using a transaction a safe update, where row will be locked, nothing will happen until the transaction is committed
		ctx := r.Context()
		var otp dtos.VerifyOTPRequest
		if err := json.NewDecoder(r.Body).Decode(&otp); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		// validate the request first
		if err := validation.Validate(&otp); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		claims, ok := r.Context().Value(middleware.UserClaimsKey).(jwt.MapClaims)
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "please login to continue")
			return
		}

		uid, ok := claims["user_id"].(float64)
		if !ok {
			utils.RespondWithError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		userID := int32(uid)
		// what can happen now is that i will be query for the hash and user_id since they make a unique key,
		// so i will get only one row, and i can check if it is expired or not, and if it is used or not, and if it is correct or not

		// start the current transaction as there is a possiblity an update would happen
		tx, err := h.DB.BeginTx(ctx, nil)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "failed to start transaction")
			return
		}
		defer tx.Rollback()
		qtx := store.New(tx)
		otp_row, err := qtx.GetValidOtpForUser(ctx, userID)
		// now first unmarshal the otp_row data, and now we need to check if the time is expired or not.
		// we already have the data in otp_row, so we can directly use it.
		// check if the otp is used or not.
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "invalid or expired OTP")
			return
		}

		// Verify the OTP hash
		if utils.ComparePassword(otp.Otpkey, otp_row.OtpKey) {
			// mark the otp as used
			_, err = qtx.MarkOtpUsed(ctx, otp_row.ID)
			if err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "failed to mark otp as used")
				return
			}

			// mark user email as verified
			_, err = qtx.MarkUserEmailVerified(ctx, userID)
			if err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "failed to mark user email as verified")
				return
			}

			if err := tx.Commit(); err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "failed to commit transaction")
				return
			}

			utils.RespondWithSuccess(w, http.StatusOK, "Email verified successfully", nil)
		} else {
			utils.RespondWithError(w, http.StatusUnauthorized, "invalid OTP code")
			return
		}
	}
}

// next steps would be assigning an api-key to make this call, because i want to rate limit, no of times
// user can call ask for verify otp, it should be like between two requests, there should be atleast 1 minute gap, and in an hour you can ask for otp
// only 5 times.
