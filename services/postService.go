package services

type PostService struct{}

var postServiceInstance *PostService = nil

// Singleton instance of PostService
func GetPostService() (*PostService, error) {
	if postServiceInstance == nil {
		postServiceInstance = &PostService{}
	}

	return postServiceInstance, nil
}

func (ps *PostService) Run() {}
