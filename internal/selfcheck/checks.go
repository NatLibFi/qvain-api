package selfcheck

import (
	"fmt"
)

type CheckFunc func() error

type Group struct {
	Name   string
	Checks []Check
}

type Check struct {
	Name string
	Func CheckFunc
}

var checks = []Group{
	{
		"basic",
		[]Check{
			{
				"check if her eyes are pretty",
				func() error {
					if false {
						return fmt.Errorf("this be a big, big error")
					}
					return nil
				},
			},
			{
				"check if she's wearing a nice dress",
				func() error {
					if false {
						return fmt.Errorf("this be a big, big error")
					}
					return nil
				},
			},
			{
				"check if her hair looks nice",
				func() error {
					if false {
						return fmt.Errorf("this be a big, big error")
					}
					return nil
				},
			},
			{
				"check if your mother called",
				func() error {
					if true {
						return fmt.Errorf("this be a big, big non-error")
					}
					return nil
				},
			},
		},
	},
	{
		"weather",
		[]Check{
			{
				"check if the weather is nice today",
				func() error {
					if false {
						return fmt.Errorf("this be a big, big error")
					}
					return nil
				},
			},
			{
				"check if it's going to rain",
				func() error {
					if false {
						return fmt.Errorf("this be a big, big error")
					}
					return nil
				},
			},
		},
	},
}
