package containerd

import (
	"context"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/image"
	"github.com/pkg/errors"
)

// TagImageWithReference adds the given reference to the image ID provided.
func (i *ImageService) TagImageWithReference(ctx context.Context, imageID image.ID, newTag reference.Named) error {
	return errdefs.NotImplemented(errors.New("not implemented"))
}
