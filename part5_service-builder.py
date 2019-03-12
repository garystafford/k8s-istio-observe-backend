#!/usr/bin/env python3
#
# author: Gary A. Stafford
# site: https://programmaticponderings.com
# license: MIT License
# purpose: Optional: (Re)build (8) Go-based microservice's Kubernetes
#          Service and Deployment resources from Jinja2 template

from jinja2 import Environment, FileSystemLoader

file_loader = FileSystemLoader('templates')
env = Environment(loader=file_loader)
template = env.get_template('service.j2')

resource_location = 'resources/services'
services = ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h']

for service in services:
    output = template.render(service=service,
                             replicas=2,
                             versions='1.2.0')
    print(output)

    filename = "service-%s%s" % (service, '.yaml')
    resource = "%s/%s" % (resource_location, filename)

    with open(resource, "w") as f:
        f.write(output)
