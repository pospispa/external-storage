# OpenShift Microversions

The `openstackmicroversions` package provides the ability to work with [OpenStack Microversions](https://developer.openstack.org/api-guide/compute/microversions.html) in Go:
- Parse OpenStack Microversions.
- Compare OpenStack Microversions.

## Installation

```
go get github.com/pospispa/openstackmicroversions
```

## Parsing OpenStack Microversion

```
microVer, err := openstackmicroversions.New("2.7")
```

## Comparision

```
package main

import (
	"fmt"

	microver "github.com/pospispa/openstackmicroversions"
)

func main() {
	var v1, v2 *microver.Microversion
	var err error
	v1, err = microver.New("2.7")
	if err != nil {
		fmt.Printf("%q", err.Error())
		return
	}
	v2, err = microver.New("3.3")
	if err != nil {
		fmt.Printf("%q", err.Error())
		return
	}
	if v1.LessThan(v2) {
		fmt.Printf("%q is smaller than %q", v1.String(), v2.String())
	} else {
		fmt.Printf("%q is greater or equal to %q", v1.String(), v2.String())
	}
}
```
