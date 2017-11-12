package ingress

import (
	"html/template"
	"net/http"
)

var defaultErrTpl = `<!doctype html>
<html>
<head>
    <title>Gourmet</title>
    <link href="https://fonts.googleapis.com/css?family=Roboto:300,400,700,900i" rel="stylesheet">
    <style>
        body {
            padding-top: 100px;
            background-color: #f7ca17;
            text-align: center;
            font-family: 'Roboto', sans-serif;
        }
        h1 {
            font-size: 200px;
            font-weight: 700;
            color: darkred;
            margin: 0;
        }
        h2 {
            font-size: 72px;
            margin: 0;
            background-color: #eee;
            padding: 10px 20px 10px 20px;
            display: inline;
            line-height: 100px;
        }
        div {
            text-align: left;
            width: 50%;
            margin: 0 auto;
            max-width: 1110px;
        }
        p {
            font-size: 22px;
            font-weight: 100;
        }
    </style>
</head>
<body>
    <div>
        <h1>{{.Status}}</h1>
        <h2>{{.StatusText}}</h2>
        <p>{{.Description}}</p>
    </div>
</body>
</html>`

const errTpl = "/etc/gourmet/tpl/error.tpl"

func writeErrTpl(w http.ResponseWriter, i interface{}) error {
	t := template.New("error-tpl")
	// TODO - Check if err tpl exists
	t, err := t.Parse(defaultErrTpl)
	if err != nil {
		return err
	}
	return t.Execute(w, i)
}
