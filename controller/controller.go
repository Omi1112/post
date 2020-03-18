package controller

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/SeijiOmi/posts-service/entity"
	"github.com/SeijiOmi/posts-service/service"
)

// Index action: GET /posts
func Index(c *gin.Context) {
	var b service.Behavior
	p, err := b.GetAllWithUserData()

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
	type tokenStru struct {
		Token string `json:"token"`
	}
	var token tokenStru
	if err := bindJSON(c, &token); err != nil {
		return
	}

	var b service.Behavior
	createdPost, err := b.CreateModel(inputPost, token.Token)

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

// HelperShow action: get /helpser/:id
func HelperShow(c *gin.Context) {
	id := c.Params.ByName("id")

	var b service.Behavior
	fmt.Println(id)
	p, err := b.GetByHelperUserIDWithUserData(id)

	if err != nil {
		c.AbortWithStatus(400)
		fmt.Println(err)
	} else {
		c.JSON(200, p)
	}
}

// SetHelpUser action: Post /helper
func SetHelpUser(c *gin.Context) {
	id, token, err := bindGetIDAndToken(c)
	if err != nil {
		return
	}

	var b service.Behavior
	p, err := b.SetHelpUserID(id, token)

	if err != nil {
		c.AbortWithStatus(400)
		fmt.Println(err)
	} else {
		c.JSON(200, p)
	}
}

// TakeHelpUser action: delete /helper
func TakeHelpUser(c *gin.Context) {
	id := c.Params.ByName("id")
	_, token, err := bindGetIDAndToken(c)
	if err != nil {
		return
	}

	var b service.Behavior
	p, err := b.TakeHelpUserID(id, token)

	if err != nil {
		c.AbortWithStatus(400)
		fmt.Println(err)
	} else {
		c.JSON(200, p)
	}
}

// DonePayment action: POST /done
func DonePayment(c *gin.Context) {
	id, token, err := bindGetIDAndToken(c)
	if err != nil {
		return
	}

	var b service.Behavior
	p, err := b.DonePayment(id, token)

	if err != nil {
		c.AbortWithStatus(400)
		fmt.Println(err)
	} else {
		c.JSON(http.StatusCreated, p)
	}
}

// DoneAcceptance action: PUT /done
func DoneAcceptance(c *gin.Context) {
	id := c.Params.ByName("id")
	_, token, err := bindGetIDAndToken(c)
	if err != nil {
		return
	}

	var b service.Behavior
	p, err := b.DoneAcceptance(id, token)

	if err != nil {
		c.AbortWithStatus(400)
		fmt.Println(err)
	} else {
		c.JSON(http.StatusCreated, p)
	}
}

// AmountPayment action: GET /amount-payment
func AmountPayment(c *gin.Context) {
	id := c.Params.ByName("id")

	var b service.Behavior
	p, err := b.GetAmountPaymentByUserID(id)

	response := struct {
		AmountPayment int
	}{
		p,
	}

	if err != nil {
		c.AbortWithStatus(400)
		fmt.Println(err)
	} else {
		c.JSON(http.StatusOK, response)
	}
}

func bindGetIDAndToken(c *gin.Context) (string, string, error) {
	type requestStru struct {
		ID    float64 `json:"id"`
		Token string  `json:"token"`
	}
	var request requestStru
	if err := bindJSON(c, &request); err != nil {
		return "", "", err
	}

	return strconv.Itoa(int(request.ID)), request.Token, nil
}

func bindJSON(c *gin.Context, data interface{}) error {
	buf := make([]byte, 2048)
	n, _ := c.Request.Body.Read(buf)
	b := string(buf[0:n])
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(b)))
	if err := c.BindJSON(data); err != nil {
		c.AbortWithStatus(400)
		fmt.Println("bind JSON err")
		fmt.Println(err)
		return err
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(b)))
	return nil
}
