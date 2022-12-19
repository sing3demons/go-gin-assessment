package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type handler struct {
	DB *sql.DB
}

type Err struct {
	Message string `json:"message"`
}

func NewApplication(db *sql.DB) *handler {
	return &handler{db}
}

func (h *handler) Greeting(c *gin.Context) {
	c.JSON(http.StatusOK, "Hello, World!")
}

type NewsExpenses struct {
	ID     int            `json:"id"`
	Title  string         `json:"title"`
	Amount float64        `json:"amount"`
	Note   string         `json:"note"`
	Tags   pq.StringArray `json:"tags"`
}

func (h *handler) CreateExpenses(c *gin.Context) {
	var m NewsExpenses

	if err := c.BindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
		return
	}

	row := h.DB.QueryRow("INSERT INTO expenses (title, amount, note, tags) VALUES ($1,$2,$3,$4) RETURNING id", m.Title, m.Amount, m.Note, m.Tags)
	err := row.Scan(&m.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, m.ID)
}

func (h *handler) ListExpenses(c *gin.Context) {
	stmt, err := h.DB.Prepare("SELECT id, title, amount, note, tags FROM expenses")
	if err != nil {
		c.JSON(http.StatusInternalServerError, Err{Message: "can't prepare query all expenses statement:" + err.Error()})
		return
	}
	rows, err := stmt.Query()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Err{Message: "can't query all expenses:" + err.Error()})
		return
	}

	//
	expenses := []NewsExpenses{}

	for rows.Next() {
		m := NewsExpenses{}
		err = rows.Scan(&m.ID, &m.Title, &m.Amount, &m.Note, &m.Tags)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Err{Message: "can't scan expenses:" + err.Error()})
			return
		}
		expenses = append(expenses, m)
	}

	c.JSON(http.StatusOK, expenses)

}

func (h *handler) GetExpensesByID(c *gin.Context) {
	var m NewsExpenses
	id := c.Param("id")
	stmt, err := h.DB.Prepare("SELECT id, title, amount, note, tags FROM expenses where id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, Err{Message: "can't prepare query expenses statement:" + err.Error()})
		return
	}
	row := stmt.QueryRow(id)

	err = row.Scan(&m.ID, &m.Title, &m.Amount, &m.Note, &m.Tags)

	switch err {
	case sql.ErrNoRows:
		c.JSON(http.StatusNotFound, Err{Message: "expenses not found"})
		return
	case nil:
		c.JSON(http.StatusOK, m)
		return
	default:
		c.JSON(http.StatusInternalServerError, Err{Message: "can't scan expenses:" + err.Error()})
		return
	}

}

func (h *handler) UpdateExpenses(c *gin.Context) {
	var m NewsExpenses

	id := c.Param("id")

	if err := c.BindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
		return
	}

	stmt, err := h.DB.Prepare("UPDATE expenses SET title=$2, amount=$3, note=$4, tags=$5 WHERE id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
		return
	}
	_, err = stmt.Exec(id, m.Title, m.Amount, m.Note, m.Tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, "Update")
}

func (h *handler) DeleteExpense(c *gin.Context) {
	id := c.Param("id")

	stmt, err := h.DB.Prepare("DELETE FROM expenses where id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
		return
	}
	_, err = stmt.Exec(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, "Delete")
}
