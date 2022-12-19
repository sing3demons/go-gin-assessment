package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/sing3demons/assessment/rest/handler"
)

func initDB() *sql.DB {
	connStr := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	createTb := `
	CREATE TABLE IF NOT EXISTS expenses (
		id SERIAL PRIMARY KEY,
		title TEXT,
		amount FLOAT,
		note TEXT,
		tags TEXT[]
	);
	`
	_, err = db.Exec(createTb)
	if err != nil {
		log.Fatal("can't create table", err)
	}

	return db
}

func basicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, password, hasAuth := c.Request.BasicAuth()
		if hasAuth && user == "admin" && password == "admin" {
			c.Next()
		} else {
			c.Abort()
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Username/Password incorrect.",
			})
			return
		}
	}
}

func main() {
	db := initDB()
	h := handler.NewApplication(db)

	r := gin.Default()
	r.GET("/", h.Greeting)

	r.POST("/expenses", h.CreateExpenses)
	r.GET("/expenses/:id", h.GetExpensesByID)
	r.PUT("/expenses/:id", h.UpdateExpenses)
	r.DELETE("/expenses/:id", h.DeleteExpense)
	// r.GET("/expenses", h.ListExpenses)
	r.GET("/expenses", basicAuth(), h.ListExpenses)

	srv := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: r,
	}
	fmt.Println("start at port:", os.Getenv("PORT"))
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown
	fmt.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("server stop")
}
