package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/SeijiOmi/posts-service/entity"
	"github.com/SeijiOmi/posts-service/service"
)

// Index action: GET /posts
func Index(c *gin.Context) {
	var b service.Behavior
	p, err := b.GetAll()

	if err != nil {
		c.AbortWithStatus(404)
		fmt.Println(err)
	} else {
		c.JSON(200, p)
	}
}

// Create action: POST /posts
func Create(c *gin.Context) {
	var inputPost entity.Post
	if err := bindJSON(c, &inputPost); err != nil {
		return
	}

	var b service.Behavior
	createdPost, err := b.CreateModel(inputPost)

	if err != nil {
		c.AbortWithStatus(400)
		fmt.Println(err)
	} else {
		c.JSON(201, createdPost)
	}
}

// Show action: GET /posts/:id
func Show(c *gin.Context) {
	id := c.Params.ByName("id")
	var b service.Behavior
	p, err := b.GetByID(id)

	if err != nil {
		c.AbortWithStatus(404)
		fmt.Println(err)
	} else {
		c.JSON(200, p)
	}
}

// Update action: PUT /posts/:id
func Update(c *gin.Context) {
	id := c.Params.ByName("id")
	var inputPost entity.Post
	if err := bindJSON(c, &inputPost); err != nil {
		return
	}

	var b service.Behavior
	p, err := b.UpdateByID(id, inputPost)

	if err != nil {
		c.AbortWithStatus(400)
		fmt.Println(err)
	} else {
		c.JSON(200, p)
	}
}

// Delete action: DELETE /posts/:id
func Delete(c *gin.Context) {
	id := c.Params.ByName("id")
	var b service.Behavior

	if err := b.DeleteByID(id); err != nil {
		c.AbortWithStatus(403)
		fmt.Println(err)
	} else {
		c.JSON(204, gin.H{"id #" + id: "deleted"})
	}
}

func bindJSON(c *gin.Context, data interface{}) error {
	if err := c.BindJSON(data); err != nil {
		c.AbortWithStatus(400)
		fmt.Println("bind JSON err")
		fmt.Println(err)
		return err
	}
	return nil
}
