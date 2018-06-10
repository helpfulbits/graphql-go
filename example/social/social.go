package social

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/helpfulbits/graphql-go/selected"
)

const Schema = `
	schema {
		query: Query
	}
	
	type Query {
		admin(id: ID!, role: Role = ADMIN): Admin!
		user(id: ID!): User!
        search(text: String!): [SearchResult]!
	}
	
	interface Admin {
		id: ID!
		name: String!
		role: Role!
	}

	scalar Time	

	type User implements Admin {
		id: ID!
		name: String!
		email: String!
		role: Role!
		phone: String!
		address: [String!]
		friends(page: Pagination): [User]
		createdAt: Time!
	}

	input Pagination {
	  first: Int
	  last: Int
	}
	
	enum Role {
		ADMIN
		USER
	}

	union SearchResult = User
`

type page struct {
	First *float64
	Last  *float64
}

type admin interface {
	IdResolver() string
	NameResolver() string
	RoleResolver() string
}

type searchResult struct {
	result interface{}
}

func (r *searchResult) ToUser() (*user, bool) {
	res, ok := r.result.(*user)
	return res, ok
}

type user struct {
	Id        string
	Name      string
	Role      string
	Email     string
	Phone     string
	Address   *[]string
	Friends   *[]*user
	CreatedAt time.Time
}

func (u user) IdResolver() string {
	return u.Id
}

func (u user) NameResolver() string {
	return u.Name
}

func (u user) RoleResolver() string {
	return u.Role
}

func (u user) FriendsResolver(args struct{ Page *page }) (*[]*user, error) {

	from := 0
	numFriends := len(*u.Friends)
	to := numFriends

	if args.Page != nil {
		if args.Page.First != nil {
			from = int(*args.Page.First)
			if from > numFriends {
				return nil, errors.New("not enough users")
			}
		}
		if args.Page.Last != nil {
			to = int(*args.Page.Last)
			if to == 0 || to > numFriends {
				to = numFriends
			}
		}
	}

	friends := (*u.Friends)[from:to]

	return &friends, nil
}

var users = []*user{
	{
		Id:        "0x01",
		Name:      "Albus Dumbledore",
		Role:      "ADMIN",
		Email:     "Albus@hogwarts.com",
		Phone:     "000-000-0000",
		Address:   &[]string{"Office @ Hogwarts", "where Horcruxes are"},
		CreatedAt: time.Now(),
	},
	{
		Id:        "0x02",
		Name:      "Harry Potter",
		Role:      "USER",
		Email:     "harry@hogwarts.com",
		Phone:     "000-000-0001",
		Address:   &[]string{"123 dorm room @ Hogwarts", "456 random place"},
		CreatedAt: time.Now(),
	},
	{
		Id:        "0x03",
		Name:      "Hermione Granger",
		Role:      "USER",
		Email:     "hermione@hogwarts.com",
		Phone:     "000-000-0011",
		Address:   &[]string{"233 dorm room @ Hogwarts", "786 @ random place"},
		CreatedAt: time.Now(),
	},
	{
		Id:        "0x04",
		Name:      "Ronald Weasley",
		Role:      "USER",
		Email:     "ronald@hogwarts.com",
		Phone:     "000-000-0111",
		Address:   &[]string{"411 dorm room @ Hogwarts", "981 @ random place"},
		CreatedAt: time.Now(),
	},
}

var usersMap = make(map[string]*user)

func init() {
	users[0].Friends = &[]*user{users[1]}
	users[1].Friends = &[]*user{users[0], users[2], users[3]}
	users[2].Friends = &[]*user{users[1], users[3]}
	users[3].Friends = &[]*user{users[1], users[2]}
	for _, usr := range users {
		usersMap[usr.Id] = usr
	}
}

type Resolver struct{}

func (r *Resolver) Admin(ctx context.Context, args struct {
	Id   string
	Role string
}) (admin, error) {
	if usr, ok := usersMap[args.Id]; ok {
		if usr.Role == args.Role {
			return *usr, nil
		}
	}
	err := fmt.Errorf("user with id=%s and role=%s does not exist", args.Id, args.Role)
	return user{}, err
}

func (r *Resolver) User(ctx context.Context, args struct{ Id string }) (user, error) {
	selectedFields, _ := selected.GetFieldsFromContext(ctx)
	for _, field := range selectedFields {
		// TODO temporary until unit test are added
		fmt.Printf("name=%s, selectedChildren=%v, args=%v\n", field.Name, len(field.Selected) > 0, field.Args)
	}

	if usr, ok := usersMap[args.Id]; ok {
		return *usr, nil
	}
	err := fmt.Errorf("user with id=%s does not exist", args.Id)
	return user{}, err
}

func (r *Resolver) Search(ctx context.Context, args struct{ Text string }) ([]*searchResult, error) {
	var result []*searchResult
	for _, usr := range users {
		if strings.Contains(usr.Name, args.Text) {
			result = append(result, &searchResult{usr})
		}
	}
	return result, nil
}
