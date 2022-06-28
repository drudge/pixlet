//go:build wasm

package encode

import (
	"fmt"
	"image"
	"syscall/js"
)

var (
	JsUint8ClampedArray js.Value = js.Global().Get("Uint8ClampedArray")
	JsImageData         js.Value = js.Global().Get("ImageData")
	jsCreateImageBitmap js.Value = js.Global().Get("createImageBitmap")
)

// Renders a screen to a sequence of JavaScript ImageBitmap objects.
func (s *Screens) EncodeImageBitmapFrames() ([]js.Value, int32, error) {
	images, err := s.render()
	if err != nil {
		return []js.Value{}, 0, err
	}

	if len(images) == 0 {
		return []js.Value{}, 0, nil
	}

	frames := make([]js.Value, 0, len(images))

	for i, img := range images {
		rgba, ok := img.(*image.RGBA)
		if !ok {
			return []js.Value{}, 0, fmt.Errorf("image %d is %T, require RGBA", i, img)
		}

		buffer := JsUint8ClampedArray.New(len(rgba.Pix))

		js.CopyBytesToJS(buffer, rgba.Pix)

		width := rgba.Bounds().Dx()
		height := rgba.Bounds().Dy()
		frame := buffer
		if !JsImageData.IsUndefined() && !jsCreateImageBitmap.IsUndefined() {
			data := JsImageData.New(buffer, width, height)
			frame = jsCreateImageBitmap.Invoke(data)
		} else {
			frame = js.ValueOf(map[string]interface{} {
				"data": buffer,
				"width": width,
				"height": height,
			})
		}
		frames = append(frames, frame)
	}

	return frames, s.delay, nil
}
