// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/chat/generate": {
            "post": {
                "description": "send message to ai",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "message"
                ],
                "summary": "send message to ai",
                "operationId": "message",
                "parameters": [
                    {
                        "description": "List of fact questions to be created",
                        "name": "ai",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/domain.NewMessage"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/domain.NewMessage"
                            }
                        }
                    }
                }
            }
        },
        "/dashboard/fillUserInfo": {
            "post": {
                "description": "User Info with the input attributes",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "User info",
                "parameters": [
                    {
                        "description": "User Info",
                        "name": "user_info",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.UserInfo"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/dto.UserInfo"
                        }
                    }
                }
            }
        },
        "/dashboard/middle/buy_premium": {
            "get": {
                "description": "buy premium user",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "buy premium",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/dashboard/middle/get-point": {
            "get": {
                "description": "get user point",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "get user point",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/dashboard/middle/get_premium": {
            "get": {
                "description": "get premium user",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "get premium",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/dashboard/showUserInfo": {
            "get": {
                "description": "Get User Info",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "User info",
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/dto.UserInfo"
                        }
                    }
                }
            }
        },
        "/fact/AnswerQuestion": {
            "get": {
                "description": "Retrieve the ID and offset from the query parameters.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "fact"
                ],
                "summary": "Get ID and Offset",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/fact/GetQuestion": {
            "get": {
                "description": "Retrieve the ID and offset from the query parameters.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "fact"
                ],
                "summary": "Get ID and Offset",
                "parameters": [
                    {
                        "type": "string",
                        "default": "\"default_id\"",
                        "description": "ID",
                        "name": "id",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "default": "\"0\"",
                        "description": "Offset",
                        "name": "offset",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dto.FactQuestions"
                        }
                    }
                }
            }
        },
        "/fact/create": {
            "post": {
                "description": "create fact",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "fact"
                ],
                "summary": "create fact",
                "operationId": "create-fact",
                "parameters": [
                    {
                        "description": "Fact",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.Fact"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/dto.Fact"
                        }
                    }
                }
            }
        },
        "/fact/createQuestions": {
            "post": {
                "description": "Creates a new fact question and returns the created fact questions.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "fact"
                ],
                "summary": "Create a fact question",
                "operationId": "create-fact-question",
                "parameters": [
                    {
                        "description": "List of fact questions to be created",
                        "name": "fact",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/dto.FactQuestions"
                            }
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/dto.FactQuestions"
                            }
                        }
                    }
                }
            }
        },
        "/fact/getFact": {
            "get": {
                "description": "Get a 5 facts",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "fact"
                ],
                "summary": "Get a fact",
                "operationId": "get-fact",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/dto.Fact"
                            }
                        }
                    }
                }
            }
        },
        "/login": {
            "post": {
                "description": "Login user with the input username,password",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Login user",
                "parameters": [
                    {
                        "description": "User",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.User"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dto.User"
                        }
                    }
                }
            }
        },
        "/news/getall": {
            "get": {
                "description": "Get all news with pagination",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "news"
                ],
                "summary": "Get all news",
                "operationId": "get-all-news",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Page number",
                        "name": "page",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dto.Response"
                        }
                    }
                }
            }
        },
        "/signup": {
            "post": {
                "description": "signup user with the input email,password",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Signup user",
                "parameters": [
                    {
                        "description": "User",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.User"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/dto.User"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "domain.NewMessage": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        },
        "dto.Choices": {
            "type": "object",
            "properties": {
                "content": {
                    "type": "string"
                },
                "is_true": {
                    "type": "boolean"
                }
            }
        },
        "dto.Fact": {
            "type": "object",
            "properties": {
                "content": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "number_of_question": {
                    "type": "integer"
                },
                "title": {
                    "type": "string"
                }
            }
        },
        "dto.FactQuestions": {
            "type": "object",
            "properties": {
                "choices": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.Choices"
                    }
                },
                "fact_id": {
                    "type": "integer"
                },
                "id": {
                    "type": "integer"
                },
                "question": {
                    "type": "string"
                }
            }
        },
        "dto.Item": {
            "type": "object",
            "properties": {
                "activity_code": {
                    "type": "string"
                },
                "activity_title": {
                    "type": "string"
                },
                "anons": {
                    "type": "string"
                },
                "anons_image": {
                    "type": "string"
                },
                "category_code": {
                    "type": "string"
                },
                "category_id": {
                    "type": "integer"
                },
                "category_title": {
                    "type": "string"
                },
                "date": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "title": {
                    "type": "string"
                },
                "url_to_web": {
                    "type": "string"
                },
                "views": {
                    "type": "integer"
                }
            }
        },
        "dto.Response": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.Item"
                    }
                }
            }
        },
        "dto.User": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "dto.UserInfo": {
            "type": "object",
            "properties": {
                "date": {
                    "type": "string"
                },
                "firstname": {
                    "type": "string"
                },
                "gender": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "lastname": {
                    "type": "string"
                },
                "skin_color": {
                    "type": "integer"
                },
                "skin_type": {
                    "type": "integer"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "web.binaryhood.uz",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Skin Ai Swagger",
	Description:      "This is a  server skin ai server.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
