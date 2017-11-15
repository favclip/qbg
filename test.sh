#!/bin/sh -eux

packages=$(go list ./...)

goapp test $packages
