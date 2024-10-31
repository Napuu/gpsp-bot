// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package openai

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/openai/openai-go/internal/apiform"
	"github.com/openai/openai-go/internal/apijson"
	"github.com/openai/openai-go/internal/param"
	"github.com/openai/openai-go/internal/requestconfig"
	"github.com/openai/openai-go/option"
)

// ImageService contains methods and other services that help with interacting with
// the openai API.
//
// Note, unlike clients, this service does not read variables from the environment
// automatically. You should not instantiate this service directly, and instead use
// the [NewImageService] method instead.
type ImageService struct {
	Options []option.RequestOption
}

// NewImageService generates a new service that applies the given options to each
// request. These options are applied after the parent client's options (if there
// is one), and before any request-specific options.
func NewImageService(opts ...option.RequestOption) (r *ImageService) {
	r = &ImageService{}
	r.Options = opts
	return
}

// Creates a variation of a given image.
func (r *ImageService) NewVariation(ctx context.Context, body ImageNewVariationParams, opts ...option.RequestOption) (res *ImagesResponse, err error) {
	opts = append(r.Options[:], opts...)
	path := "images/variations"
	err = requestconfig.ExecuteNewRequest(ctx, http.MethodPost, path, body, &res, opts...)
	return
}

// Creates an edited or extended image given an original image and a prompt.
func (r *ImageService) Edit(ctx context.Context, body ImageEditParams, opts ...option.RequestOption) (res *ImagesResponse, err error) {
	opts = append(r.Options[:], opts...)
	path := "images/edits"
	err = requestconfig.ExecuteNewRequest(ctx, http.MethodPost, path, body, &res, opts...)
	return
}

// Creates an image given a prompt.
func (r *ImageService) Generate(ctx context.Context, body ImageGenerateParams, opts ...option.RequestOption) (res *ImagesResponse, err error) {
	opts = append(r.Options[:], opts...)
	path := "images/generations"
	err = requestconfig.ExecuteNewRequest(ctx, http.MethodPost, path, body, &res, opts...)
	return
}

// Represents the url or the content of an image generated by the OpenAI API.
type Image struct {
	// The base64-encoded JSON of the generated image, if `response_format` is
	// `b64_json`.
	B64JSON string `json:"b64_json"`
	// The prompt that was used to generate the image, if there was any revision to the
	// prompt.
	RevisedPrompt string `json:"revised_prompt"`
	// The URL of the generated image, if `response_format` is `url` (default).
	URL  string    `json:"url"`
	JSON imageJSON `json:"-"`
}

// imageJSON contains the JSON metadata for the struct [Image]
type imageJSON struct {
	B64JSON       apijson.Field
	RevisedPrompt apijson.Field
	URL           apijson.Field
	raw           string
	ExtraFields   map[string]apijson.Field
}

func (r *Image) UnmarshalJSON(data []byte) (err error) {
	return apijson.UnmarshalRoot(data, r)
}

func (r imageJSON) RawJSON() string {
	return r.raw
}

type ImageModel = string

const (
	ImageModelDallE2 ImageModel = "dall-e-2"
	ImageModelDallE3 ImageModel = "dall-e-3"
)

type ImagesResponse struct {
	Created int64              `json:"created,required"`
	Data    []Image            `json:"data,required"`
	JSON    imagesResponseJSON `json:"-"`
}

// imagesResponseJSON contains the JSON metadata for the struct [ImagesResponse]
type imagesResponseJSON struct {
	Created     apijson.Field
	Data        apijson.Field
	raw         string
	ExtraFields map[string]apijson.Field
}

func (r *ImagesResponse) UnmarshalJSON(data []byte) (err error) {
	return apijson.UnmarshalRoot(data, r)
}

func (r imagesResponseJSON) RawJSON() string {
	return r.raw
}

type ImageNewVariationParams struct {
	// The image to use as the basis for the variation(s). Must be a valid PNG file,
	// less than 4MB, and square.
	Image param.Field[io.Reader] `json:"image,required" format:"binary"`
	// The model to use for image generation. Only `dall-e-2` is supported at this
	// time.
	Model param.Field[ImageModel] `json:"model"`
	// The number of images to generate. Must be between 1 and 10. For `dall-e-3`, only
	// `n=1` is supported.
	N param.Field[int64] `json:"n"`
	// The format in which the generated images are returned. Must be one of `url` or
	// `b64_json`. URLs are only valid for 60 minutes after the image has been
	// generated.
	ResponseFormat param.Field[ImageNewVariationParamsResponseFormat] `json:"response_format"`
	// The size of the generated images. Must be one of `256x256`, `512x512`, or
	// `1024x1024`.
	Size param.Field[ImageNewVariationParamsSize] `json:"size"`
	// A unique identifier representing your end-user, which can help OpenAI to monitor
	// and detect abuse.
	// [Learn more](https://platform.openai.com/docs/guides/safety-best-practices/end-user-ids).
	User param.Field[string] `json:"user"`
}

func (r ImageNewVariationParams) MarshalMultipart() (data []byte, contentType string, err error) {
	buf := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(buf)
	err = apiform.MarshalRoot(r, writer)
	if err != nil {
		writer.Close()
		return nil, "", err
	}
	err = writer.Close()
	if err != nil {
		return nil, "", err
	}
	return buf.Bytes(), writer.FormDataContentType(), nil
}

// The format in which the generated images are returned. Must be one of `url` or
// `b64_json`. URLs are only valid for 60 minutes after the image has been
// generated.
type ImageNewVariationParamsResponseFormat string

const (
	ImageNewVariationParamsResponseFormatURL     ImageNewVariationParamsResponseFormat = "url"
	ImageNewVariationParamsResponseFormatB64JSON ImageNewVariationParamsResponseFormat = "b64_json"
)

func (r ImageNewVariationParamsResponseFormat) IsKnown() bool {
	switch r {
	case ImageNewVariationParamsResponseFormatURL, ImageNewVariationParamsResponseFormatB64JSON:
		return true
	}
	return false
}

// The size of the generated images. Must be one of `256x256`, `512x512`, or
// `1024x1024`.
type ImageNewVariationParamsSize string

const (
	ImageNewVariationParamsSize256x256   ImageNewVariationParamsSize = "256x256"
	ImageNewVariationParamsSize512x512   ImageNewVariationParamsSize = "512x512"
	ImageNewVariationParamsSize1024x1024 ImageNewVariationParamsSize = "1024x1024"
)

func (r ImageNewVariationParamsSize) IsKnown() bool {
	switch r {
	case ImageNewVariationParamsSize256x256, ImageNewVariationParamsSize512x512, ImageNewVariationParamsSize1024x1024:
		return true
	}
	return false
}

type ImageEditParams struct {
	// The image to edit. Must be a valid PNG file, less than 4MB, and square. If mask
	// is not provided, image must have transparency, which will be used as the mask.
	Image param.Field[io.Reader] `json:"image,required" format:"binary"`
	// A text description of the desired image(s). The maximum length is 1000
	// characters.
	Prompt param.Field[string] `json:"prompt,required"`
	// An additional image whose fully transparent areas (e.g. where alpha is zero)
	// indicate where `image` should be edited. Must be a valid PNG file, less than
	// 4MB, and have the same dimensions as `image`.
	Mask param.Field[io.Reader] `json:"mask" format:"binary"`
	// The model to use for image generation. Only `dall-e-2` is supported at this
	// time.
	Model param.Field[ImageModel] `json:"model"`
	// The number of images to generate. Must be between 1 and 10.
	N param.Field[int64] `json:"n"`
	// The format in which the generated images are returned. Must be one of `url` or
	// `b64_json`. URLs are only valid for 60 minutes after the image has been
	// generated.
	ResponseFormat param.Field[ImageEditParamsResponseFormat] `json:"response_format"`
	// The size of the generated images. Must be one of `256x256`, `512x512`, or
	// `1024x1024`.
	Size param.Field[ImageEditParamsSize] `json:"size"`
	// A unique identifier representing your end-user, which can help OpenAI to monitor
	// and detect abuse.
	// [Learn more](https://platform.openai.com/docs/guides/safety-best-practices/end-user-ids).
	User param.Field[string] `json:"user"`
}

func (r ImageEditParams) MarshalMultipart() (data []byte, contentType string, err error) {
	buf := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(buf)
	err = apiform.MarshalRoot(r, writer)
	if err != nil {
		writer.Close()
		return nil, "", err
	}
	err = writer.Close()
	if err != nil {
		return nil, "", err
	}
	return buf.Bytes(), writer.FormDataContentType(), nil
}

// The format in which the generated images are returned. Must be one of `url` or
// `b64_json`. URLs are only valid for 60 minutes after the image has been
// generated.
type ImageEditParamsResponseFormat string

const (
	ImageEditParamsResponseFormatURL     ImageEditParamsResponseFormat = "url"
	ImageEditParamsResponseFormatB64JSON ImageEditParamsResponseFormat = "b64_json"
)

func (r ImageEditParamsResponseFormat) IsKnown() bool {
	switch r {
	case ImageEditParamsResponseFormatURL, ImageEditParamsResponseFormatB64JSON:
		return true
	}
	return false
}

// The size of the generated images. Must be one of `256x256`, `512x512`, or
// `1024x1024`.
type ImageEditParamsSize string

const (
	ImageEditParamsSize256x256   ImageEditParamsSize = "256x256"
	ImageEditParamsSize512x512   ImageEditParamsSize = "512x512"
	ImageEditParamsSize1024x1024 ImageEditParamsSize = "1024x1024"
)

func (r ImageEditParamsSize) IsKnown() bool {
	switch r {
	case ImageEditParamsSize256x256, ImageEditParamsSize512x512, ImageEditParamsSize1024x1024:
		return true
	}
	return false
}

type ImageGenerateParams struct {
	// A text description of the desired image(s). The maximum length is 1000
	// characters for `dall-e-2` and 4000 characters for `dall-e-3`.
	Prompt param.Field[string] `json:"prompt,required"`
	// The model to use for image generation.
	Model param.Field[ImageModel] `json:"model"`
	// The number of images to generate. Must be between 1 and 10. For `dall-e-3`, only
	// `n=1` is supported.
	N param.Field[int64] `json:"n"`
	// The quality of the image that will be generated. `hd` creates images with finer
	// details and greater consistency across the image. This param is only supported
	// for `dall-e-3`.
	Quality param.Field[ImageGenerateParamsQuality] `json:"quality"`
	// The format in which the generated images are returned. Must be one of `url` or
	// `b64_json`. URLs are only valid for 60 minutes after the image has been
	// generated.
	ResponseFormat param.Field[ImageGenerateParamsResponseFormat] `json:"response_format"`
	// The size of the generated images. Must be one of `256x256`, `512x512`, or
	// `1024x1024` for `dall-e-2`. Must be one of `1024x1024`, `1792x1024`, or
	// `1024x1792` for `dall-e-3` models.
	Size param.Field[ImageGenerateParamsSize] `json:"size"`
	// The style of the generated images. Must be one of `vivid` or `natural`. Vivid
	// causes the model to lean towards generating hyper-real and dramatic images.
	// Natural causes the model to produce more natural, less hyper-real looking
	// images. This param is only supported for `dall-e-3`.
	Style param.Field[ImageGenerateParamsStyle] `json:"style"`
	// A unique identifier representing your end-user, which can help OpenAI to monitor
	// and detect abuse.
	// [Learn more](https://platform.openai.com/docs/guides/safety-best-practices/end-user-ids).
	User param.Field[string] `json:"user"`
}

func (r ImageGenerateParams) MarshalJSON() (data []byte, err error) {
	return apijson.MarshalRoot(r)
}

// The quality of the image that will be generated. `hd` creates images with finer
// details and greater consistency across the image. This param is only supported
// for `dall-e-3`.
type ImageGenerateParamsQuality string

const (
	ImageGenerateParamsQualityStandard ImageGenerateParamsQuality = "standard"
	ImageGenerateParamsQualityHD       ImageGenerateParamsQuality = "hd"
)

func (r ImageGenerateParamsQuality) IsKnown() bool {
	switch r {
	case ImageGenerateParamsQualityStandard, ImageGenerateParamsQualityHD:
		return true
	}
	return false
}

// The format in which the generated images are returned. Must be one of `url` or
// `b64_json`. URLs are only valid for 60 minutes after the image has been
// generated.
type ImageGenerateParamsResponseFormat string

const (
	ImageGenerateParamsResponseFormatURL     ImageGenerateParamsResponseFormat = "url"
	ImageGenerateParamsResponseFormatB64JSON ImageGenerateParamsResponseFormat = "b64_json"
)

func (r ImageGenerateParamsResponseFormat) IsKnown() bool {
	switch r {
	case ImageGenerateParamsResponseFormatURL, ImageGenerateParamsResponseFormatB64JSON:
		return true
	}
	return false
}

// The size of the generated images. Must be one of `256x256`, `512x512`, or
// `1024x1024` for `dall-e-2`. Must be one of `1024x1024`, `1792x1024`, or
// `1024x1792` for `dall-e-3` models.
type ImageGenerateParamsSize string

const (
	ImageGenerateParamsSize256x256   ImageGenerateParamsSize = "256x256"
	ImageGenerateParamsSize512x512   ImageGenerateParamsSize = "512x512"
	ImageGenerateParamsSize1024x1024 ImageGenerateParamsSize = "1024x1024"
	ImageGenerateParamsSize1792x1024 ImageGenerateParamsSize = "1792x1024"
	ImageGenerateParamsSize1024x1792 ImageGenerateParamsSize = "1024x1792"
)

func (r ImageGenerateParamsSize) IsKnown() bool {
	switch r {
	case ImageGenerateParamsSize256x256, ImageGenerateParamsSize512x512, ImageGenerateParamsSize1024x1024, ImageGenerateParamsSize1792x1024, ImageGenerateParamsSize1024x1792:
		return true
	}
	return false
}

// The style of the generated images. Must be one of `vivid` or `natural`. Vivid
// causes the model to lean towards generating hyper-real and dramatic images.
// Natural causes the model to produce more natural, less hyper-real looking
// images. This param is only supported for `dall-e-3`.
type ImageGenerateParamsStyle string

const (
	ImageGenerateParamsStyleVivid   ImageGenerateParamsStyle = "vivid"
	ImageGenerateParamsStyleNatural ImageGenerateParamsStyle = "natural"
)

func (r ImageGenerateParamsStyle) IsKnown() bool {
	switch r {
	case ImageGenerateParamsStyleVivid, ImageGenerateParamsStyleNatural:
		return true
	}
	return false
}
