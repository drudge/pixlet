package main

import (
	"fmt"
	"syscall/js"

	"tidbyt.dev/pixlet/encode"
	"tidbyt.dev/pixlet/runtime"
)

const magnify = 1.0

func JsRunApplet(code []byte) (*encode.Screens, error) {
	applet := runtime.Applet{}
	err := applet.Load("code.star", code, nil)
	if err != nil {
		return nil, err
	}

	roots, err := applet.Run(map[string]string{})
	if err != nil {
		return nil, err
	}

	return encode.ScreensFromRoots(roots), nil
}

func JsRender(this js.Value, args []js.Value) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Invalid number of arguments!")
	}

	code := []byte(args[0].String())

	screens, err := JsRunApplet(code)
	if err != nil {
		return nil, err
	}

	frames, delay, err := screens.EncodeImageBitmapFrames()
	if err != nil {
		return nil, err
	}

	a := js.Global().Get("Array").New(len(frames))

	for i, frame := range frames {
		a.SetIndex(i, frame)
	}

	result := js.ValueOf(map[string]interface{}{
		"frames": a,
		"delay":  delay,
	})

	return result, nil
}

type JsFunction func(this js.Value, args []js.Value) (interface{}, error)

var (
	JsError   js.Value = js.Global().Get("Error")
	JsPromise js.Value = js.Global().Get("Promise")
)

func JsCreateAsyncFunction(innerFunc JsFunction) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		handler := js.FuncOf(func(_ js.Value, promFn []js.Value) interface{} {
			resolve, reject := promFn[0], promFn[1]

			go func() {
				defer func() {
					if r := recover(); r != nil {
						reject.Invoke(JsError.New(fmt.Sprintf("%v", r)))
					}
				}()

				res, err := innerFunc(this, args)
				if err != nil {
					reject.Invoke(JsError.New(err.Error()))
				} else {
					resolve.Invoke(res)
				}
			}()

			return nil
		})

		return JsPromise.New(handler)
	})
}

func main() {
	done := make(chan bool)
	runtime.InitCache(runtime.NewInMemoryCache())

	js.Global().Set("pixlet", js.ValueOf(map[string]interface{} {
		"render": JsCreateAsyncFunction(JsRender),
	}))
	<-done
}
