package postgres

import (
	"database/sql"
	"fmt"

	// postgres driver
	_ "github.com/lib/pq"
)

// Db is our database struct used for interacting with the database
type Db struct {
	*sql.DB
}

// New makes a new database using the connection string and
// returns it, otherwise returns the error
func New(connString string) (*Db, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	// Check that our connection is good
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &Db{db}, nil
}

// ConnString returns a connection string based on the parameters it's given
// This would normally also contain the password, however we're not using one
func ConnString(host string, port int, user string, dbName string) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s sslmode=disable",
		host, port, user, dbName,
	)
}

// User shape
type Person struct {
	ID        int
	Name      string
	Height    int
	Mass      int
	Gender    string
	Homeworld string
}

// GetPeople is called within our user query for graphql
func (d *Db) GetPeople(name string) []Person {
	// Prepare query, takes a name argument, protects from sql injection
	stmt, err := d.Prepare("SELECT * FROM people WHERE name=$1")
	if err != nil {
		fmt.Println("GetPeople Preperation Err: ", err)
	}

	// Make query with our stmt, passing in name argument
	rows, err := stmt.Query(name)
	if err != nil {
		fmt.Println("GetPeople Query Err: ", err)
	}

	// Create User struct for holding each row's data
	var r Person
	// Create slice of Users for our response
	people := []Person{}
	// Copy the columns from row into the values pointed at by r (User)
	for rows.Next() {
		err = rows.Scan(
			&r.Name,
			&r.Height,
			&r.Mass,
			&r.Gender,
			&r.Homeworld,
		)
		if err != nil {
			fmt.Println("Error scanning rows: ", err)
		}
		people = append(people, r)
	}

	return people
}
