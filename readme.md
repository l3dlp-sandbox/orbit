# Orbit &middot; ![prerelease](https://img.shields.io/badge/project-Pre--Release-red) ![codecov](https://img.shields.io/codecov/c/github/guyaross/orbit) [![CodeFactor](https://www.codefactor.io/repository/github/guyaross/orbit/badge)](https://www.codefactor.io/repository/github/guyaross/orbit) [![GitHub license](https://img.shields.io/badge/license-GNU_GPLv3-blue.svg)](./LICENSE) 
Orbit is a golang server side processing framework for building server side web applications.

- **Micro-frontend**: Out of the box support for React and vanilla JavaScript micro frontends.
- **Static bundling**: Automatically creates static HTML files for components that don't make use of server side processing. 
- **Bundling support**: Orbit currently has support for the following tools:

| Name               | Extent of support |
|--------------------|-------------------|
| Vanilla Javascript | Full support      |
| Client side React  | Full support      |
| Server side React  | Experimental      |
| Vue                | Planned           |



## Installation
- **Manual Installation**: To install manually, clone this repo and compile it with `go build`.

## Examples
There are several examples exist in the [./examples](/examples). Here is a basic one to get you up and running.

1. Initialize the workspace directory with `orbit init`, then follow the prompts
2. Create a react component
```jsx
// /pages/hello-world.jsx
const HelloWorldComponent = ({ from }) => {
    return (
        <>
            <div> Hello, from {from} </div>
        </>
    )
}

export default HelloWorldComponent
```
3. Run the build process with `orbit build`
4. Create golang application
```go
// main.go
package main

import (
    // ... 
)

func main() {
    // orbitgen comes from an autogenerated lib output from orbit
    orb, err := orbitgen.New()
    if err != nil {
        panic(err)
    }

    orb.HandleFunc("/", func (r *orbitgen.Request) {
        props := make(map[string]interface{})
        // sets the prop 'from' to the string 'orbit'
        props["from"] = "orbit"

        // renders a single page & passes props into the component
        c.RenderPage(orbitgen.HelloWorldComponent, props)

        // can also use c.RenderPages(...) to make build a micro-frontend
    })

    http.ListenAndServe(":3030", orb.Serve())
}

```
5. Run golang application with `go run main.go`

## Contributing

### [Contributing Guide](./CONTRIBUTING.md)
Please first read our [contributing guide](./CONTRIBUTING.md) before contributing to this project.

### [Good First Issues](https://github.com/GuyARoss/orbit/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)
To gain exposure to the project you can find a list of [good first issues](https://github.com/GuyARoss/orbit/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).

### License
Orbit it licensed under [GNU GPLv3](./LICENSE) 