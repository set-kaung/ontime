package service

// func TestInserUser(t *testing.T) {
// 	f, err := os.OpenFile("users.json", os.O_RDWR, 0777)
// 	if err != nil {
// 		t.Log("test file not found")
// 		t.Fail()
// 		return
// 	}
// 	mockRepo := repository.NewMockUserRepo(f, &sync.RWMutex{})
// 	userService := &service.UserService{Repo: mockRepo}
// 	email, _ := domain.NewEmail("user456@gmail.com")
// 	err = userService.InsertUser(email, "user456", "pass")
// 	if err != nil {
// 		t.Errorf("%v\n", err)
// 	}
// }

// func TestInserUserUserAlreadyExist(t *testing.T) {
// 	f, err := os.OpenFile("users.json", os.O_RDWR, 0777)
// 	if err != nil {
// 		t.Log("test file not found")
// 		t.Fail()
// 		return
// 	}
// 	mockRepo := repository.NewMockUserRepo(f, &sync.RWMutex{})
// 	userService := &service.UserService{Repo: mockRepo}
// 	email, _ := domain.NewEmail("user456@gmail.com")
// 	err = userService.InsertUser(email, "user456", "pass")
// 	if err != nil {
// 		t.Errorf("%v\n", err)
// 	}
// }
