package main

import (
	"bytes"
	"fmt"
	"strings"
	"vezdekod-chat-bots/types"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
)

func newMessageWithImages(vk *api.VK, user *types.User, message string, images []*types.Image) (*params.MessagesSendBuilder, error) {
	b := params.NewMessagesSendBuilder()
	attachs := []string{}
	for _, image := range images {
		resp, err := vk.UploadMessagesPhoto(user.ID, bytes.NewReader(image.ImgBytes))
		if err != nil {
			return nil, err
		}
		photo := resp[0]
		attachs = append(attachs, fmt.Sprintf("photo%d_%d", photo.OwnerID, photo.ID))
	}

	b.Message(message)
	b.Attachment(strings.Join(attachs, ","))
	b.RandomID(0)
	b.PeerID(user.ID)

	return b, nil
}
