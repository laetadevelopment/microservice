package v1

import (
	"context"

	"github.com/golang/protobuf/ptypes"
	"github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/laetadevelopment/microservice/microservices/template/pkg/api/v1"
)

const (
	apiVersion = "v1"
)

type templateServiceServer struct {
	db *mongo.Client
}

// MongoRepository implementation
type MongoRepository struct {
	collection *mongo.Collection
}

// NewTemplateServiceServer connects to MongoDB
func NewTemplateServiceServer(db *mongo.Client) v1.TemplateServiceServer {
	return &templateServiceServer{db: db}
}

func (s *templateServiceServer) checkAPI(api string) error {
	if len(api) > 0 {
		if apiVersion != api {
			return status.Errorf(codes.Unimplemented,
				"unsupported API version: service implements API version '%s', but asked for '%s'", apiVersion, api)
		}
	}
	return nil
}

// Create a new template in MongoDB
func (s *templateServiceServer) Create(ctx context.Context, req *v1.CreateRequest) (*v1.CreateResponse, error) {
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	c := s.db.Database("template").Collection("template")

	templateId := uuid.NewV1().String()

	p := v1.Template{
		Id:          templateId,
		Items:		 req.Template.Items,
		Created:     ptypes.TimestampNow(),
		Updated:     ptypes.TimestampNow(),
	}

	_, err := c.InsertOne(ctx, &p)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to insert into Template.template-> "+err.Error())
	}

	return &v1.CreateResponse{
		Api: apiVersion,
		Id: templateId,
	}, nil
}

// Read a template in MongoDB
func (s *templateServiceServer) Read(ctx context.Context, req *v1.ReadRequest) (*v1.ReadResponse, error) {
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	c := s.db.Database("template").Collection("template")

	filter := bson.D{{"id", req.Id}}

	var p v1.Template

	err := c.FindOne(context.TODO(), filter).Decode(&p)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to find document-> "+err.Error())
	}

	return &v1.ReadResponse{
		Api:     apiVersion,
		Template: &p,
	}, nil

}

// Update a template in MongoDB
func (s *templateServiceServer) Update(ctx context.Context, req *v1.UpdateRequest) (*v1.UpdateResponse, error) {
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	c := s.db.Database("template").Collection("template")

	filter := bson.D{{"id", req.Template.Id}}
	update := bson.D{
		{"$set", bson.D{
			{"items", req.Template.Items},
			{"updated", ptypes.TimestampNow()},
		}},
	}

	updateResult, err := c.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to update document-> "+err.Error())
	}

	u := updateResult.ModifiedCount

	return &v1.UpdateResponse{
		Api:     apiVersion,
		Updated: u,
	}, nil
}

// Delete a template in MongoDB
func (s *templateServiceServer) Delete(ctx context.Context, req *v1.DeleteRequest) (*v1.DeleteResponse, error) {
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	c := s.db.Database("template").Collection("template")

	filter := bson.D{{"id", req.Id}}

	deleteResult, err := c.DeleteOne(context.TODO(), filter)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to delete document-> "+err.Error())
	}

	d := deleteResult.DeletedCount

	return &v1.DeleteResponse{
		Api:     apiVersion,
		Deleted: d,
	}, nil
}

// List all templates available via MongoDB Client
func (s *templateServiceServer) List(ctx context.Context, req *v1.ListRequest) (*v1.ListResponse, error) {
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	c := s.db.Database("template").Collection("template")

	findOptions := options.Find()
	var list []*v1.Template

	cur, err := c.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to find documents in Template.template-> "+err.Error())
	}

	for cur.Next(context.TODO()) {
		var elem v1.Template
		err := cur.Decode(&elem)
		if err != nil {
			return nil, status.Error(codes.Unknown, "failed to decode document-> "+err.Error())
		}

		list = append(list, &elem)
	}

	if err := cur.Err(); err != nil {
		return nil, status.Error(codes.Unknown, "failed reading documents-> "+err.Error())
	}

	cur.Close(context.TODO())

	return &v1.ListResponse{
		Api:  apiVersion,
		Data: list,
	}, nil
}
