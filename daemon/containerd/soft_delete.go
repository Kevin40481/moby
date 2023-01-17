package containerd

import (
	"context"

	cerrdefs "github.com/containerd/containerd/errdefs"
	containerdimages "github.com/containerd/containerd/images"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/errdefs"
	"github.com/sirupsen/logrus"
)

// softImageDelete deletes the image, making sure that there are other images
// that reference the content of the deleted image.
// If no other image exists, a dangling one is created.
func (i *ImageService) softImageDelete(img containerdimages.Image) error {
	// Don't accept ctx as argument as we don't want this
	// to be interrupted in the middle.
	ctx := context.Background()
	is := i.client.ImageService()

	logger := logrus.WithFields(logrus.Fields{
		"name":   img.Name,
		"digest": img.Target.Digest.String(),
	})

	ref, err := reference.ParseNormalizedNamed(img.Name)
	if err != nil {
		logger.WithError(err).Debug("image name is not a valid reference")
		return errdefs.InvalidParameter(err)
	}

	// If the image already exists, persist it as dangling image
	// but only if no other image has the same target.
	digest := img.Target.Digest.String()
	imgs, err := is.List(ctx, "target.digest=="+digest)
	if err != nil {
		logger.WithField("digest", digest).WithError(err).Debug("failed to check if there are images targeting this digest")
		return errdefs.System(err)
	}

	// Create dangling image if this is the last image pointing to this target.
	if len(imgs) == 1 {
		danglingImage := img

		refWithoutTag := reference.TrimNamed(ref)
		newRef, err := reference.WithDigest(refWithoutTag, img.Target.Digest)
		if err != nil {
			logger.WithError(err).Error("could not create a digested reference for dangling image")
			return errdefs.Unknown(err)
		}

		danglingImage.Name = newRef.String()
		delete(danglingImage.Labels, "io.containerd.image.name")
		delete(danglingImage.Labels, "org.opencontainers.image.ref.name")

		_, err = is.Create(ctx, danglingImage)

		// Error out in case we couldn't persist the old image.
		// If it already exists, then just continue.
		if err != nil && !cerrdefs.IsAlreadyExists(err) {
			logger.WithError(err).Error("failed to create a dangling image for the image being replaced")
			return errdefs.System(err)
		}
	}

	// Free the target name.
	err = is.Delete(ctx, img.Name)
	if err != nil {
		if !cerrdefs.IsNotFound(err) {
			logger.WithError(err).Error("failed to delete image which existed a moment before")
			return errdefs.System(err)
		}
	}

	return nil
}
