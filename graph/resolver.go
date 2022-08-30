package graph

import "github.com/campbelljlowman/fazool-api/graph/model"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{
	sessions map[int]*model.Session
}
