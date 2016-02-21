/*
Copyright 2014 go-trello authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package trello

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"
)

type Board struct {
	client   *Client
	Id       string `json:"id"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	DescData struct {
		Emoji struct{} `json:"emoji"`
	} `json:"descData"`
	Closed         bool   `json:"closed"`
	IdOrganization string `json:"idOrganization"`
	Pinned         bool   `json:"pinned"`
	Url            string `json:"url"`
	ShortUrl       string `json:"shortUrl"`
	Prefs          struct {
		PermissionLevel       string            `json:"permissionLevel"`
		Voting                string            `json:"voting"`
		Comments              string            `json:"comments"`
		Invitations           string            `json:"invitations"`
		SelfJoin              bool              `json:"selfjoin"`
		CardCovers            bool              `json:"cardCovers"`
		CardAging             string            `json:"cardAging"`
		CalendarFeedEnabled   bool              `json:"calendarFeedEnabled"`
		Background            string            `json:"background"`
		BackgroundColor       string            `json:"backgroundColor"`
		BackgroundImage       string            `json:"backgroundImage"`
		BackgroundImageScaled []BoardBackground `json:"backgroundImageScaled"`
		BackgroundTile        bool              `json:"backgroundTile"`
		BackgroundBrightness  string            `json:"backgroundBrightness"`
		CanBePublic           bool              `json:"canBePublic"`
		CanBeOrg              bool              `json:"canBeOrg"`
		CanBePrivate          bool              `json:"canBePrivate"`
		CanInvite             bool              `json:"canInvite"`
	} `json:"prefs"`
	LabelNames struct {
		Red    string `json:"red"`
		Orange string `json:"orange"`
		Yellow string `json:"yellow"`
		Green  string `json:"green"`
		Blue   string `json:"blue"`
		Purple string `json:"purple"`
	} `json:"labelNames"`
}

type BoardBackground struct {
	width  int    `json:"width"`
	height int    `json:"height"`
	url    string `json:"url"`
}

func (c *Client) Boards() (boards []Board, err error) {
	body, err := c.Get("/boards/")
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &boards)
	for i := range boards {
		boards[i].client = c
	}
	return
}

func (c *Client) Board(boardId string) (board *Board, err error) {
	body, err := c.Get("/boards/" + boardId)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &board)
	board.client = c
	return
}

func (b *Board) Lists() (lists []List, err error) {
	body, err := b.client.Get("/boards/" + b.Id + "/lists")
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &lists)
	for i := range lists {
		lists[i].client = b.client
	}
	return
}

func (b *Board) Members() (members []Member, err error) {
	body, err := b.client.Get("/boards/" + b.Id + "/members")
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &members)
	for i := range members {
		members[i].client = b.client
	}
	return
}

func (b *Board) Cards() (cards []Card, err error) {
	body, err := b.client.Get("/boards/" + b.Id + "/cards")
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &cards)
	for i := range cards {
		cards[i].client = b.client
	}
	return
}

func (b *Board) Card(IdCard string) (card *Card, err error) {
	body, err := b.client.Get("/boards/" + b.Id + "/cards/" + IdCard)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &card)
	card.client = b.client
	return
}

func (b *Board) Checklists() (checklists []Checklist, err error) {
	body, err := b.client.Get("/boards/" + b.Id + "/checklists")
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &checklists)
	for i := range checklists {
		checklists[i].client = b.client
	}
	return
}

func (b *Board) MemberCards(IdMember string) (cards []Card, err error) {
	body, err := b.client.Get("/boards/" + b.Id + "/members/" + IdMember + "/cards")
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &cards)
	for i := range cards {
		cards[i].client = b.client
	}
	return
}

func (b *Board) Actions() (actions []Action, err error) {
	body, err := b.client.Get("/boards/" + b.Id + "/actions")
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &actions)
	for i := range actions {
		actions[i].client = b.client
	}
	return
}

type AddCardOpts struct {
	Name        string
	Description string
	Position    string
	Due         *time.Time
	ListID      string
	Labels      []string
	Members     []string
}

func (a AddCardOpts) validate() (bool, error) {
	if len(a.Name) < 1 || len(a.Name) > 16384 {
		return false, errors.New("Name must be a string of length from 1 to 16384")
	}

	if len(a.Description) > 16384 {
		return false, errors.New("Description may not be longer than 16384 characters")
	}

	if a.Position != "" && (a.Position != "bottom" && a.Position != "top") {
		return false, errors.New("If position is present it has to be 'bottom' or 'top'")
	}

	//TODO: Maybe it's not a good idea to hardcode the number 24 even if the documentation is explicit about it
	if len(a.ListID) != 24 {
		return false, errors.New("ListID is required and must be a valid 24-char hex string")
	}

	return true, nil
}

func (b *Board) AddCard(opts AddCardOpts) (*Card, error) {
	if ok, err := opts.validate(); !ok {
		return nil, err
	}

	params := url.Values{
		"name":      []string{opts.Name},
		"idList":    []string{opts.ListID},
		"urlSource": []string{"null"}, // Not yet implemented
	}

	if len(opts.Description) > 0 {
		params.Set("desc", opts.Description)
	}

	if len(opts.Position) > 0 {
		params.Set("pos", opts.Position)
	}

	if len(opts.Labels) > 0 {
		params.Set("idLabels", strings.Join(opts.Labels, ","))
	}

	if len(opts.Members) > 0 {
		params.Set("idMembers", strings.Join(opts.Members, ","))
	}

	if opts.Due == nil {
		params.Set("due", "null")
	} else {
		params.Set("due", opts.Due.Format("2006-01-02T15:04:05-07:00"))
	}

	resp, err := b.client.Post("/cards", params)
	if err != nil {
		return nil, err
	}

	c := &Card{client: b.client}
	return c, json.Unmarshal(resp, c)
}
