package main

import (
	"context"
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

func StartGraphQLServer(service *BenchmarkService) *http.Server {
	responseType := graphql.NewObject(graphql.ObjectConfig{
		Name: "BenchmarkResponse",
		Fields: graphql.Fields{
			"transport":       &graphql.Field{Type: graphql.String},
			"message":         &graphql.Field{Type: graphql.String},
			"request_time":    &graphql.Field{Type: graphql.String},
			"response_time":   &graphql.Field{Type: graphql.String},
			"total_time_ms":   &graphql.Field{Type: graphql.Int},
			"current_load":    &graphql.Field{Type: graphql.Int},
			"fastest_hint":    &graphql.Field{Type: graphql.String},
			"processed_value": &graphql.Field{Type: graphql.String},
		},
	})

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"health": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return "ok", nil
				},
			},
		},
	})

	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"process": &graphql.Field{
				Type: responseType,
				Args: graphql.FieldConfigArgument{
					"message": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"work_ms": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					message, _ := p.Args["message"].(string)

					var workMS int32
					if val, ok := p.Args["work_ms"].(int); ok {
						workMS = int32(val)
					}

					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					return service.Handle(ctx, "graphql", BenchmarkRequest{
						Message: message,
						WorkMS:  workMS,
					})
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
	if err != nil {
		panic(err)
	}

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	return &http.Server{
		Addr:    ":8080",
		Handler: h,
	}
}
