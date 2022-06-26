package controllers

import (
	"ambassador/src/database"
	"ambassador/src/models"
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"sort"
	"strconv"
	"strings"
	"time"
)

func Products(c *fiber.Ctx) error {
	var products []models.Product

	database.DB.Find(&products)

	return c.JSON(products)
}

func ProductsFrontend(c *fiber.Ctx) error {
	var products []models.Product
	var ctx = context.Background()

	result, err := database.Cache.Get(ctx, "products_frontend").Result()

	if err != nil {
		database.DB.Find(&products)

		productsInBytes, err := json.Marshal(products)
		if err != nil {
			panic(err)
		}

		if errorRedisSet := database.Cache.Set(ctx, "products_frontend", productsInBytes, 30*time.Minute).Err(); errorRedisSet != nil {
			panic(errorRedisSet)
		}

	} else {
		err := json.Unmarshal([]byte(result), &products)
		if err != nil {
			panic(err)
		}
	}
	return c.JSON(products)
}

func ProductBackend(c *fiber.Ctx) error {
	perPage := 9
	var products []models.Product
	var sortedProducts []models.Product
	var data []models.Product
	var ctx = context.Background()

	result, err := database.Cache.Get(ctx, "products_backend").Result()

	if err != nil {
		database.DB.Find(&products)

		productsInBytes, err := json.Marshal(products)
		if err != nil {
			panic(err)
		}
		if errorRedisSet := database.Cache.Set(ctx, "products_backend", productsInBytes, 30*time.Minute).Err(); errorRedisSet != nil {
			panic(errorRedisSet)
		}
	} else {
		err := json.Unmarshal([]byte(result), &products)
		if err != nil {
			panic(err)
		}
	}

	if searchParam := c.Query("search"); searchParam != "" {
		for _, product := range products {
			formattedSearchParam := strings.ToLower(searchParam)
			formattedName := strings.ToLower(product.Title)
			formattedDesc := strings.ToLower(product.Description)
			if strings.Contains(formattedName, formattedSearchParam) || strings.Contains(formattedDesc, formattedSearchParam) {
				sortedProducts = append(sortedProducts, product)
			}
		}
	} else {
		sortedProducts = products
	}

	if sortParam := c.Query("sort"); sortParam != "" {
		formattedSortParam := strings.ToLower(sortParam)
		if formattedSortParam == "asc" {
			sort.Slice(sortedProducts, func(i, j int) bool {
				return sortedProducts[i].Price < sortedProducts[j].Price
			})
		} else if formattedSortParam == "desc" {
			sort.Slice(sortedProducts, func(i, j int) bool {
				return sortedProducts[i].Price > sortedProducts[j].Price
			})
		}
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page <= 0 {
		page = 1
	}

	var productsAmount int = len(sortedProducts)
	var outOfRange bool = page*perPage >= productsAmount && productsAmount >= (page-1)*perPage
	var validRange bool = productsAmount >= page*perPage

	if outOfRange {
		data = sortedProducts[perPage*(page-1) : productsAmount]
	} else if validRange {
		data = sortedProducts[perPage*(page-1) : perPage*page]
	} else {
		data = []models.Product{}
	}

	return c.JSON(fiber.Map{
		"data":      data,
		"total":     productsAmount,
		"page":      page,
		"last_page": productsAmount/perPage + 1,
	})

}

func CreateProduct(c *fiber.Ctx) error {
	var product models.Product

	if err := c.BodyParser(&product); err != nil {
		return err
	}

	database.DB.Create(&product)

	return c.JSON(product)
}

func GetProduct(c *fiber.Ctx) error {
	var product models.Product

	id, _ := strconv.Atoi(c.Params("id"))

	product.Id = uint(id)

	database.DB.Find(&product)

	return c.JSON(product)
}

func UpdateProduct(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	product := models.Product{}
	product.Id = uint(id)

	if err := c.BodyParser(&product); err != nil {
		return err
	}
	database.DB.Model(&product).Updates(&product)

	return c.JSON(product)
}

func DeleteProduct(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	product := models.Product{}
	product.Id = uint(id)

	database.DB.Delete(&product)

	return nil
}
