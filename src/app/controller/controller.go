package controller

import (
	"app/config/db"
	"app/model"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

//User sign up
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user model.User
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &user)
	var res model.ResponseResult
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	collection, err := db.GetDBCollection("User", "users")
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	var result model.User
	err = collection.FindOne(context.TODO(), bson.D{{"email", user.Email}}).Decode(&result)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 5)

			if err != nil {
				res.Error = "Error While Hashing Password, Try Again"
				json.NewEncoder(w).Encode(res)
				return
			}
			user.Password = string(hash)
			_, err = collection.InsertOne(context.TODO(), user)
			if err != nil {
				res.Error = "Error While Creating User, Try Again"
				json.NewEncoder(w).Encode(res)
				return
			}
			res.Result = "Registration Successful"
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	res.Result = "Email already Exists!!"
	json.NewEncoder(w).Encode(res)
	return
}

//User log in
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user model.User
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &user)
	if err != nil {
		log.Fatal(err)
	}
	collection, err := db.GetDBCollection("User", "users")
	if err != nil {
		log.Fatal(err)
	}
	var result model.User
	var res model.ResponseResult
	err = collection.FindOne(context.TODO(), bson.D{{"email", user.Email}}).Decode(&result)
	if err != nil {
		res.Error = "Invalid Email"
		json.NewEncoder(w).Encode(res)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user.Password))
	if err != nil {
		res.Error = "Invalid password"
		json.NewEncoder(w).Encode(res)
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":     result.Email,
		"firstname": result.FirstName,
		"lastname":  result.LastName,
	})
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		res.Error = "Error while generating token,Try again"
		json.NewEncoder(w).Encode(res)
		return
	}
	result.Token = tokenString
	json.NewEncoder(w).Encode(result)

}

//Gets posts from external API and saves them in the database
func SavePostsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query().Get("q")
	var posts []model.Post
	response, err := http.Get("https://invidio.us/api/v1/search?q=" + q)
	if err != nil {
		println(err)
	}
	body, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(body, &posts)
	var data []interface{}
	for _, t := range posts {
		t.ID = primitive.NewObjectID()
		t.CreatedAt = time.Now()
		t.UpdatedAt = time.Now()
		data = append(data, t)
	}
	var res model.ResponseResult
	collection, err := db.GetDBCollection("Post", "posts")
	_, err = collection.InsertMany(context.TODO(), data)
	if err != nil {
		res.Error = "Error While Saving posts, Try Again"
		json.NewEncoder(w).Encode(res)
		return
	}
	res.Result = "Posts Successfully Saved"
	json.NewEncoder(w).Encode(data)
	return
}

//Gets posts from database and displays them with pagination
func GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	count := r.URL.Query().Get("count")
	p := r.URL.Query().Get("page")
	limit, err := strconv.ParseInt(count, 0, 64)
	page, err := strconv.ParseInt(p, 0, 64)
	skip := (page - 1) * limit
	opts := options.FindOptions{
		Skip:  &skip,
		Limit: &limit,
	}
	var results []*model.Post
	collection, err := db.GetDBCollection("Post", "posts")
	cur, err := collection.Find(context.Background(), bson.D{{}}, &opts)
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(context.Background()) {
		var post model.Post
		err := cur.Decode(&post)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, &post)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(results)
	cur.Close(context.TODO())
}

//Creates a new post
func CreatePostHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var post model.Post
	_ = json.NewDecoder(request.Body).Decode(&post)
	collection, err := db.GetDBCollection("Post", "posts")
	var res model.ResponseResult
	post.ID = primitive.NewObjectID()
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(response).Encode(res)
		return
	}
	result, _ := collection.InsertOne(context.Background(), &post)
	json.NewEncoder(response).Encode(result)
}

//Gets one post from database by id
func GetPostHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := request.URL.Query().Get("id")
	id, _ := primitive.ObjectIDFromHex(params)
	collection, err := db.GetDBCollection("Post", "posts")
	var res model.ResponseResult
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(response).Encode(res)
		return
	}
	findResult := collection.FindOne(context.Background(), bson.M{"_id": &id})
	var post model.Post
	err = findResult.Decode(&post)
	if err != nil {
		json.NewEncoder(response).Encode("Post does not exist")
		fmt.Println(err)
		return
	}
	json.NewEncoder(response).Encode(post)
}

//Updates a post
func UpdatePostHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := request.URL.Query().Get("id")
	var post model.Post
	_ = json.NewDecoder(request.Body).Decode(&post)
	id, _ := primitive.ObjectIDFromHex(params)
	collection, err := db.GetDBCollection("Post", "posts")
	resultUpdate, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"title":      post.Title,
				"authore":    post.Author,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	json.NewEncoder(response).Encode(resultUpdate)
}

// Deletes a post
func DeletePostHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := request.URL.Query().Get("id")
	id, _ := primitive.ObjectIDFromHex(params)
	collection, err := db.GetDBCollection("Post", "posts")
	resultDelete, err := collection.DeleteOne(context.Background(), bson.M{"_id": &id})
	if err != nil {
		fmt.Println(err)
		return
	}
	json.NewEncoder(response).Encode(resultDelete)
}
