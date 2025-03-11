package mdencoder

import (
	"fmt"
	"testing"
)

// Profile is a nested struct inside User.
type Profile struct {
	Age  *int   `json:"age,omitempty" jsonschema:"title=Age,description=User's age"`
	Bio  string `json:"bio,omitempty" jsonschema:"title=Bio"`
	City string `json:"city,omitempty"`
}

// Tag represents an object inside the Tags array.
type Tag struct {
	Category string              `json:"category,omitempty" jsonschema:"title=Category,description=Type of tag"`
	Tag      string              `json:"tag,omitempty" jsonschema:"title=Tag,description=Actual tag value"`
	Exps     []string            `json:"exps,omitempty" jsonschema:"title=Exps,description=Extend preferences"`
	Map      []map[string]string `json:"map,omitempty" jsonschema:"title=Map,description=Test for map"`
}

// User represents a sample struct with `jsonschema` tags.
type User struct {
	ID      int                 `json:"id" jsonschema:"title=User ID,description=The unique identifier for a user"`
	Name    *string             `json:"name,omitempty" jsonschema:"title=Full Name,description=The full name of the user"`
	Email   *string             `json:"email,omitempty" jsonschema:"title=Email,description=The email address of the user"`
	Profile *Profile            `json:"profile,omitempty" jsonschema:"title=Profile,description=User profile information"`
	Tags    []*Tag              `json:"tags,omitempty" jsonschema:"title=Tags,description=List of user tags"`
	Map     map[string][]string `json:"map,omitempty" jsonschema:"title=Map,description=Test for map"`
}

// CustomProfile is a nested struct inside User.
type CustomProfile struct {
	Age  *int   `json:"age,omitempty" jsonschema:"title=Age,description=User's age"`
	Bio  string `json:"bio,omitempty" jsonschema:"title=Bio"`
	City string `json:"city,omitempty"`
}

func (p CustomProfile) MarshalMarkdown() ([]byte, error) {
	return []byte("Custom marshaler for profile"), nil
}

// User represents a sample struct with `jsonschema` tags.
type CustomUser struct {
	ID      int                 `json:"id" jsonschema:"title=User ID,description=The unique identifier for a user"`
	Name    *string             `json:"name,omitempty" jsonschema:"title=Full Name,description=The full name of the user"`
	Email   *string             `json:"email,omitempty" jsonschema:"title=Email,description=The email address of the user"`
	Profile *CustomProfile      `json:"profile,omitempty" jsonschema:"title=Profile,description=User profile information"`
	Tags    []*Tag              `json:"tags,omitempty" jsonschema:"title=Tags,description=List of user tags"`
	Map     map[string][]string `json:"map,omitempty" jsonschema:"title=Map,description=Test for map"`
}

var (
	name  = "Alice"
	email = "alice@example.com"
	age   = 30

	user = User{
		ID:    123,
		Name:  &name,
		Email: &email,
		Profile: &Profile{
			Age:  &age,
			Bio:  "Loves coding and coffee.",
			City: "Bangkok",
		},
		Tags: []*Tag{
			{Category: "Interest", Tag: "Technology", Map: []map[string]string{
				{"key1": "v1"},
				{"key2": "v2"},
			}},
			nil, // Simulating a nil pointer in the slice
			{Category: "Skill", Tag: "Golang", Exps: []string{"tag1", "tag2", "tag3"}},
		},
		Map: map[string][]string{
			"key1": {"v1", "v1.2"},
			"key2": {"v2", "v2.2"},
		},
	}

	customUser = CustomUser{
		ID:    123,
		Name:  &name,
		Email: &email,
		Profile: &CustomProfile{
			Age:  &age,
			Bio:  "Loves coding and coffee.",
			City: "Bangkok",
		},
		Tags: []*Tag{
			{Category: "Interest", Tag: "Technology", Map: []map[string]string{
				{"key1": "v1"},
				{"key2": "v2"},
			}},
			nil, // Simulating a nil pointer in the slice
			{Category: "Skill", Tag: "Golang", Exps: []string{"tag1", "tag2", "tag3"}},
		},
		Map: map[string][]string{
			"key1": {"v1", "v1.2"},
			"key2": {"v2", "v2.2"},
		},
	}
)

func TestEncoder(t *testing.T) {
	// Example object with pointers and an array of object pointers

	markdown, err := Encode(&user)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(string(markdown))
	// Output:
	// # **User ID**: 123
	//
	// _The unique identifier for a user_
	//
	// # **Full Name**: Alice
	//
	// _The full name of the user_
	//
	// # **Email**: alice@example.com
	//
	// _The email address of the user_
	//
	//
	// # **Profile**: User profile information
	// - **Age**: 30
	//
	//   _User's age_
	//
	// - **Bio**: Loves coding and coffee.
	//
	// - **city**: Bangkok
	//
	//
	// # **Tags**: List of user tags
	// - **Category**: Interest
	//
	//   _Type of tag_
	//
	//   **Tag**: Technology
	//
	//   _Actual tag value_
	//
	//
	//   **Map**: Test for map
	//   - **key1**: v1
	//
	//   - **key2**: v2
	//
	// - **Category**: Skill
	//
	//   _Type of tag_
	//
	//   **Tag**: Golang
	//
	//   _Actual tag value_
	//
	//
	//   **Exps**: Extend preferences
	//   - tag1
	//   - tag2
	//   - tag3
	//
	// # **Map**: Test for map
	// - **key1**
	//   - v1
	//   - v1.2
	// - **key2**
	//   - v2
	//   - v2.2
}

func TestMarshaler(t *testing.T) {
	markdown, err := Encode(&customUser)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(string(markdown))
	// Output:
	// # User ID (The unique identifier for a user)
	// 123
	//
	// # Full Name (The full name of the user)
	// Alice
	//
	// # Email (The email address of the user)
	// alice@example.com
	//
	// # Profile (User profile information)
	// Custom marshaler for profile
	//
	// # Tags (List of user tags)
	// - Category (Type of tag): Interest
	//
	//   Tag (Actual tag value): Technology
	//
	//   Map (Test for map)
	//   - key1: v1
	//   - key2: v2
	//
	// - Category (Type of tag): Skill
	//
	//   Tag (Actual tag value): Golang
	//
	//   Exps (Extend preferences)
	//   - tag1
	//   - tag2
	//   - tag3
	//
	//
	// # Map (Test for map)
	// - key1
	//   - v1
	//   - v1.2
	// - key2
	//   - v2
	//   - v2.2
}
