package containerd

import (
	"context"

	cerrdefs "github.com/containerd/containerd/errdefs"
	containerdimages "github.com/containerd/containerd/images"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/image"
	"github.com/sirupsen/logrus"
)

// TagImageWithReference creates an image named as newTag and targeting the given descriptor id.
func (i *ImageService) TagImageWithReference(ctx context.Context, imageID image.ID, newTag reference.Named) error {
	logger := logrus.WithFields(logrus.Fields{
		"imageID": imageID.String(),
		"tag":     newTag.String(),
	})

	target, err := i.resolveDescriptor(ctx, imageID.String())
	if err != nil {
		logger.WithError(err).Debug("failed to resolve image id to a descriptor")
		return err
	}

	img := containerdimages.Image{
		Name:   newTag.String(),
		Target: target,
	}

	is := i.client.ImageService()
	_, err = is.Create(ctx, img)
	if err != nil {
		if !cerrdefs.IsAlreadyExists(err) {
			logger.WithError(err).Error("failed to create image")
			return errdefs.System(err)
		}

		// If there already exists an image with this tag, delete it
		err := is.Delete(ctx, img.Name)

		if err != nil {
			logger.WithError(err).Error("failed to delete old image")
			return errdefs.System(err)
		}

		if _, err = is.Create(ctx, img); err != nil {
			logger.WithError(err).Error("failed to create an image after deleting old image")
			return errdefs.System(err)
		}
	}

	logger.Info("image created")
	return nil
}
