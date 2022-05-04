/*
Copyright AppsCode Inc.

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

package server

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/go-macaron/cache"
	"google.golang.org/api/youtube/v3"
	"gopkg.in/macaron.v1"
)

const youtubeChannelID = "UCxObRDZ0DtaQe_cCP-dN-xg"

func (s *Server) RegisterYoutubeAPI(m *macaron.Macaron) {
	m.Get("/_/playlists", func(ctx *macaron.Context, c cache.Cache, log *log.Logger) {
		key := ctx.Req.URL.Path
		out := c.Get(key)
		if out == nil {
			lists, err := s.ListPlaylists(youtubeChannelID)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			out = lists
			_ = c.Put(key, out, 60) // cache for 60 seconds
		} else {
			log.Println(key, "found")
		}
		ctx.JSON(http.StatusOK, out)
	})

	m.Get("/_/playlists/:id", func(ctx *macaron.Context, c cache.Cache, log *log.Logger) {
		key := ctx.Req.URL.Path
		out := c.Get(key)
		if out == nil {
			playlistID := ctx.Params("id")
			items, err := s.ListPlaylistItems(playlistID)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			out = items
			_ = c.Put(key, out, 60) // cache for 60 seconds
		} else {
			log.Println(key, "found")
		}
		ctx.JSON(http.StatusOK, out)
	})
}

// https://developers.google.com/youtube/v3/guides/implementation/playlists
func (s *Server) ListPlaylists(channelID string) ([]*youtube.Playlist, error) {
	call := s.srvYT.Playlists.List(strings.Split("snippet,contentDetails,status", ","))
	// call = call.Fields("items(id,snippet(title,description,publishedAt,tags,thumbnails(high)),contentDetails,status)")
	call = call.ChannelId(channelID)

	var out []*youtube.Playlist
	err := call.Pages(context.TODO(), func(resp *youtube.PlaylistListResponse) error {
		out = append(out, resp.Items...)
		return nil
	})
	return out, err
}

// https://developers.google.com/youtube/v3/docs/playlistItems/list
func (s *Server) ListPlaylistItems(playlistID string) ([]*youtube.PlaylistItem, error) {
	call := s.srvYT.PlaylistItems.List(strings.Split("snippet,contentDetails,status", ","))
	call = call.Fields("items(snippet(title,description,position,thumbnails(high)),contentDetails,status)")
	call = call.PlaylistId(playlistID)

	var out []*youtube.PlaylistItem
	err := call.Pages(context.Background(), func(resp *youtube.PlaylistItemListResponse) error {
		out = append(out, resp.Items...)
		return nil
	})
	return out, err
}
