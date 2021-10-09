/*
	Packages and libs
*/

package main

//Libs
import (
    "net/http"
    "github.com/gin-gonic/gin"
	"fmt"
	"os"
	"io/ioutil"
	"encoding/json"
	"context"
	"time"
	
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    //"go.mongodb.org/mongo-driver/mongo/readpref"
)


//Structure for inserting data in db
type Users struct {
    Users []User `json:"users"`
}
type User struct {
	UID	   string `json:"uid"`
    Name   string `json:"name"`
    Email  string `json:"email"`
    PWD    string  `json:"pwd"`  
}

// Structure for new post
type Posts struct {
    Posts []Post `json:"posts"`
}
type Post struct {
	PID	      string `json:"pid"`
	UID		  string `json:"uid"`
    Caption   string `json:"caption"`
    URL       string `json:"url"`
    Time      string  `json:"time"`  
}

/*
	Functions used in db
*/

// This method closes mongoDB connection and cancel context.
func close(client *mongo.Client, ctx context.Context,
	cancel context.CancelFunc){
	 
defer cancel()

defer func() {
 if err := client.Disconnect(ctx); err != nil {
	 panic(err)
 }
}()
}

// This is a user defined method that returns mongo.Client,
func connect(uri string)(*mongo.Client, context.Context, context.CancelFunc, error) {
 
    ctx, cancel := context.WithTimeout(context.Background(),30 * time.Second)
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    return client, ctx, cancel, err
}

// Insertion of record
func insertOne(client *mongo.Client, ctx context.Context, dataBase, col string, doc interface{})(*mongo.InsertOneResult, error) {
 
    collection := client.Database(dataBase).Collection(col) 
    result, err := collection.InsertOne(ctx, doc)
    return result, err
}

// Retriving the record
func query(client *mongo.Client, ctx context.Context,dataBase, col string, query, field interface{})(result *mongo.Cursor, err error) {
		
		collection := client.Database(dataBase).Collection(col)
		result, err = collection.Find(ctx, query,
									  options.Find().SetProjection(field))
		return
	}


// Main function

func  main()  {

	router := gin.Default()

	router.GET("/users/:id", getUser)
    router.POST("/users", newUser)
	router.GET("/posts/:id", getPost)
	router.POST("/posts", newPost)
	router.GET("/posts/users/:id", allUsersPost)

    router.Run("localhost:8080")
}


/*
	Function for Routes
*/


// Get user details
func getUser(c *gin.Context) {

	userid := c.Param("id")

	client, ctx, cancel, err := connect("mongodb://localhost:27017")
    if err != nil {
        panic(err)
    }
	defer close(client, ctx, cancel)
	var filter, option interface{}
	filter = bson.D{
        {"uid", bson.D{{"$eq", userid}}},
    }
	option = bson.D{{"_id", 0}}
	cursor, err := query(client, ctx, "appointy","user", filter, option)
	if err != nil {
        panic(err)
    }
 
    var results []bson.D					 
	if err := cursor.All(ctx, &results); err != nil {
     
        // handle the error
        panic(err)
    }
     
    // printing the result of query.
    fmt.Println("Query Reult")
    for _, doc := range results {
        fmt.Println(doc)
		c.IndentedJSON(http.StatusOK, doc)
    }
   
}

//Insert new User
func newUser(c *gin.Context) {
    
	jsonFile, err := os.Open("user.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
    var users Users
    json.Unmarshal(byteValue, &users)
	client, ctx, cancel, err := connect("mongodb://localhost:27017")
    if err != nil {
        panic(err)
    }
	defer close(client, ctx, cancel)
	var document interface{}
	document = bson.D{
        {"uid", users.Users[0].UID},
        {"name", users.Users[0].Name},
        {"email", users.Users[0].Email},
        {"password", users.Users[0].PWD},
    }

	insertOneResult, err := insertOne(client, ctx, "appointy", "user", document)
     
    // handle the error
    if err != nil {
        panic(err)
    }
	fmt.Println("Result of InsertOne")
    fmt.Println(insertOneResult.InsertedID)


    c.IndentedJSON(http.StatusCreated, gin.H{
                "status": "success",
				"users": document,
	})	
}

//Get post

func getPost(c *gin.Context){
	postid := c.Param("id")

	client, ctx, cancel, err := connect("mongodb://localhost:27017")
    if err != nil {
        panic(err)
    }
	defer close(client, ctx, cancel)
	var filter, option interface{}
	filter = bson.D{
        {"pid", bson.D{{"$eq", postid}}},
    }
	option = bson.D{{"_id", 0}}
	cursor, err := query(client, ctx, "appointy","posts", filter, option)
	if err != nil {
        panic(err)
    }
 
    var results []bson.D					 
	if err := cursor.All(ctx, &results); err != nil {
     
        // handle the error
        panic(err)
    }
     
    // printing the result of query.
    fmt.Println("Query Reult")
    for _, doc := range results {
        fmt.Println(doc)
		c.IndentedJSON(http.StatusOK, doc)
    }
}

//Insert new post
func newPost(c *gin.Context)  {
	jsonFile, err := os.Open("post.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
    var posts Posts
    json.Unmarshal(byteValue, &posts)
	client, ctx, cancel, err := connect("mongodb://localhost:27017")
    if err != nil {
        panic(err)
    }
	defer close(client, ctx, cancel)
	var document interface{}
	document = bson.D{
        {"pid", posts.Posts[0].PID},
        {"uid", posts.Posts[0].UID},
        {"caption", posts.Posts[0].Caption},
        {"url", posts.Posts[0].URL},
		{"time", posts.Posts[0].Time},
    }

	insertOneResult, err := insertOne(client, ctx, "appointy", "posts", document)
     
    // handle the error
    if err != nil {
        panic(err)
    }
	fmt.Println("Result of InsertOne")
    fmt.Println(insertOneResult.InsertedID)


    c.IndentedJSON(http.StatusCreated, gin.H{
                "status": "success",
				"posts": document,
	})
}

// Get all posts of an user

func allUsersPost(c *gin.Context)  {
	usertid := c.Param("id")

	client, ctx, cancel, err := connect("mongodb://localhost:27017")
    if err != nil {
        panic(err)
    }
	defer close(client, ctx, cancel)
	var filter, option interface{}
	filter = bson.D{
        {"uid", bson.D{{"$eq", usertid}}},
    }
	option = bson.D{{"_id", 0}}
	cursor, err := query(client, ctx, "appointy","posts", filter, option)
	if err != nil {
        panic(err)
    }
 
    var results []bson.D					 
	if err := cursor.All(ctx, &results); err != nil {
     
        // handle the error
        panic(err)
    }
     
    // printing the result of query.
    fmt.Println("Query Reult")
    for _, doc := range results {
        fmt.Println(doc)
		c.IndentedJSON(http.StatusOK, doc)
    }
}