//go:build !custom || processors || processors.aws_ec2

package all

import _ "Dana/plugins/processors/aws_ec2" // register plugin
