package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var secretKey = []byte("secretKey")

type Book struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	ISBN       string    `json:"isbn"`
	Author     string    `json:"author"`
	Price      float64   `json:"price"`
	CoverImage string    `json:"cover_image"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

var books = []Book{
	{ID: 1, Title: "Book 1", ISBN: "1234567890", Author: "Author 1", Price: 29.99, CoverImage: "cover1.png", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: 2, Title: "Book 2", ISBN: "0987654321", Author: "Author 2", Price: 39.99, CoverImage: "cover2.png", CreatedAt: time.Now(), UpdatedAt: time.Now()},
}

func main() {
	router := gin.Default()

	router.POST("/login", login)
	authGroup := router.Group("/api")
	authGroup.Use(authMiddleware())
	{
		authGroup.GET("/books", getBooks)
		authGroup.GET("/books/:id", getBookByID)
		authGroup.POST("/books", createBook)
		authGroup.PUT("/books/:id", updateBook)
		authGroup.DELETE("/books/:id", deleteBook)
	}

	router.Run(":8080")
}

func login(c *gin.Context) {
	var user struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if user.Username == "user" && user.Password == "password" {
		token, err := generateToken(user.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getBooks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"books": books})
}

func getBookByID(c *gin.Context) {
	id := getIDParam(c)
	book, err := findBookByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"book": book})
}

func createBook(c *gin.Context) {
	var newBook Book
	if err := c.ShouldBindJSON(&newBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	newBook.ID = generateID()
	newBook.CreatedAt = time.Now()
	newBook.UpdatedAt = time.Now()

	books = append(books, newBook)

	c.JSON(http.StatusCreated, gin.H{"book": newBook})
}

func updateBook(c *gin.Context) {
	id := getIDParam(c)
	bookIndex, err := findBookIndexByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	var updatedBook Book
	if err := c.ShouldBindJSON(&updatedBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	updatedBook.ID = id
	updatedBook.CreatedAt = books[bookIndex].CreatedAt
	updatedBook.UpdatedAt = time.Now()

	books[bookIndex] = updatedBook
	c.JSON(http.StatusOK, gin.H{"book": updatedBook})
}

func deleteBook(c *gin.Context) {
	id := getIDParam(c)
	bookIndex, err := findBookIndexByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	deletedBook := books[bookIndex]
	books = append(books[:bookIndex], books[bookIndex+1:]...)

	c.JSON(http.StatusOK, gin.H{"book": deletedBook})
}

func getIDParam(c *gin.Context) int {
	id, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		// Handle error, misalnya return nilai default atau memberikan respon error
		return 0
	}
	return id
}

func generateID() int {
	return len(books) + 1
}

func findBookByID(id int) (Book, error) {
	for _, book := range books {
		if book.ID == id {
			return book, nil
		}
	}
	return Book{}, fmt.Errorf("Book not found")
}

func findBookIndexByID(id int) (int, error) {
	for i, book := range books {
		if book.ID == id {
			return i, nil
		}
	}
	return -1, fmt.Errorf("Book not found")
}

func generateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 1).Unix(),
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
