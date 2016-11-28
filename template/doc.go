/*

Package sockaddr/template provides a text/template interface the SockAddr helper
functions.  The primary entry point into the sockaddr/template package is
through its Parse() call.  For example:

    import (
      "fmt"

      template "github.com/hashicorp/go-sockaddr/template"
    )

    results, err := template.Parse(`{{ GetPrivateIP }}`)
    if err != nil {
      fmt.Errorf("Unable to find a private IP address: %v", err)
    }
    fmt.Printf("My Private IP address is: %s\n", results)

Below is a list of builtin template functions and details re: their usage.  It
is possible to add additional functions by calling ParseIfAddrsTemplate
directly.

In general, the calling convention for this template library is to seed a list
of initial interfaces via one of the Get*Interfaces() calls, then filter, sort,
and extract the necessary attributes for use as string input.  This template
interface is primarily geared toward resolving specific values that are only
available at runtime, but can be defined as a heuristic for execution when a
config file is parsed.

All functions, unless noted otherwise, return an array of IfAddr structs making
it possible to `sort`, `filter`, `limit`, seek (via the `offset` function), or
`unique` the list.  To extract useful string information, the `attr` and `join`
functions return a single string value.  See below for details.

Important note: see the
https://github.com/hashicorp/go-sockaddr/tree/master/cmd/sockaddr utility for
more examples and for a CLI utility to experiment with the template syntax.

`GetAllInterfaces` - Returns an exhaustive set of IfAddr structs available on
the host.  `GetAllInterfaces` is the initial input and accessible as the initial
"dot" in the pipeline.

Example:

    {{ GetAllInterfaces }}


`GetDefaultInterfaces` - Returns one IfAddr for every IP that is on the
interface containing the default route for the host.

Example:

    {{ GetDefaultInterfaces }}

`GetPrivateInterfaces` - Returns one IfAddr for every IP that matches RFC 6890
and attached to the interface with the default route.  NOTE: RFC 6890 is a more
exhaustive version of RFC1918 because it spans IPv4 and IPv6, however it does
permit the inclusion of likely undesired addresses such as multicast, therefore
it may be prudent to use this in conjunction with additional filtering.

Example:

    {{ GetPrivateInterfaces | include "flags" "forwardable" }}


`GetPublicInterfaces` - Returns a list of IfAddr that do not match RFC 6890 and
is attached to the default route.

Example:

    {{ GetPublicInterfaces | include "flags" "forwardable" }}


`GetPrivateIP` - Helper function that returns a string of the private IP address
(RFC 6890) that is attached to the default route.

Example:

    {{ GetPrivateIP }}


`GetPublicIP` - Helper function that returns a string of the public IP (non-RFC
6890) that is attached to the default route.

Example:

    {{ GetPublicIP }}


`sort` - Sorts the IfAddrs result based on its arguments.  `sort` takes one
argument, a list of ways to sort its IfAddrs argument.  The list of sort
criteria is comma separated (`,`):
  - `address`, `+address`: Ascending sort of IfAddrs by Address
  - `-address`: Descending sort of IfAddrs by Address
  - `name`, `+name`: Ascending sort of IfAddrs by lexical ordering of interface name
  - `-name`: Descending sort of IfAddrs by lexical ordering of interface name
  - `port`, `+port`: Ascending sort of IfAddrs by port number
  - `-port`: Descending sort of IfAddrs by port number
  - `private`, `+private`: Ascending sort of IfAddrs with private addresses first
  - `-private`: Descending sort IfAddrs with private addresses last
  - `size`, `+size`: Ascending sort of IfAddrs by their network size as determined
    by their netmask (larger networks first)
  - `-size`: Descending sort of IfAddrs by their network size as determined by their
    netmask (smaller networks first)
  - `type`, `+type`: Ascending sort of IfAddrs by the type of the IfAddr (Unix,
    IPv4, then IPv6)
  - `-type`: Descending sort of IfAddrs by the type of the IfAddr (IPv6, IPv4, Unix)

Example:

    {{ GetPrivateInterfaces | sort "type,size,address" }}


`exclude` and `include`: Filters IfAddrs based on the selector criteria and its
arguments.  Both `exclude` and `include` take two arguments.  The list of
available filtering criteria is:
  - "address": Filter IfAddrs based on a regexp matching the string representation
    of the address
  - "flag","flags": Filter IfAddrs based on the list of flags specified.  Multiple
    flags can be passed together using the pipe character (`|`) to create an inclusive
    bitmask of flags.  The list of flags is included below.
  - "name": Filter IfAddrs based on a regexp matching the interface name.
  - "port": Filter IfAddrs based on an exact match of the port number (number must
    be expressed as a string)
  - "rfc", "rfcs": Filter IfAddrs based on the matching RFC.  If more than one RFC
    is specified, the list of RFCs can be joined together using the pipe character (`|`).
  - "size": Filter IfAddrs based on the exact match of the mask size.
  - "type": Filter IfAddrs based on their SockAddr type.  Multiple types can be
    specified together by using the pipe character (`|`).  Valid types include:
    `ip`, `ipv4`, `ipv6`, and `unix`.

Example:

    {{ GetPrivateInterfaces | exclude "type" "IPv6" | include "flag" "up|forwardable" }}


`unique`: Removes duplicate entries from the IfAddrs list, assuming the list has
already been sorted.  `unique` only takes one argument:
  - "address": Removes duplicates with the same address
  - "name": Removes duplicates with the same interface names

Example:

    {{ GetPrivateInterfaces | sort "type,address" | unique "name" }}


`limit`: Reduces the size of the list to the specified value.

Example:

    {{ GetPrivateInterfaces | include "flags" "forwardable|up" | limit 1 }}


`offset`: Seeks into the list by the specified value.  A negative value can be
used to seek from the end of the list.

Example:

    {{ GetPrivateInterfaces | include "flags" "forwardable|up" | offset "-2" | limit 1 }}


`attr`: Extracts a single attribute of the first member of the list and returns
it as a string.  `attr` takes a single attribute name.  The list of available
attributes is type-specific and shared between `join`.  See below for a list of
supported attributes.

Example:

    {{ GetPrivateInterfaces | include "flags" "forwardable|up" | attr "address" }}


`join`: Similar to `attr`, `join` extracts all matching attributes of the list
and returns them as a string joined by the separator, the second argument to
`join`.  The list of available attributes is type-specific and shared between
`join`.

Example:

    {{ GetPrivateInterfaces | include "flags" "forwardable|up" | join "address" " " }}


`exclude` and `include` flags:
  - `broadcast`
  - `down`: Is the interface down?
  - `forwardable`: Is the IP forwardable?
  - `global unicast`
  - `interface-local multicast`
  - `link-local multicast`
  - `link-local unicast`
  - `loopback`
  - `multicast`
  - `point-to-point`
  - `unspecified`: Is the IfAddr the IPv6 unspecified address?
  - `up`: Is the interface up?


Attributes for `attr` and `join`:

SockAddr Type:
  - `string`
  - `type`

IPAddr Type:
  - `address`
  - `binary`
  - `first_usable`
  - `hex`
  - `host`
  - `last_usable`
  - `mask_bits`
  - `netmask`
  - `network`
  - `octets`: Decimal values per byte
  - `port`
  - `size`: Number of hosts in the network

IPv4Addr Type:
  - `broadcast`
  - `uint32`: unsigned integer representation of the value

IPv6Addr Type:
  - `uint128`: unsigned integer representation of the value

UnixSock Type:
  - `path`

*/
package template