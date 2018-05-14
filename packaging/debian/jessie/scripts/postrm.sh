#!/bin/sh
/bin/systemctl stop spqr
/bin/systemctl disable spqr
/bin/systemctl daemon-reload
