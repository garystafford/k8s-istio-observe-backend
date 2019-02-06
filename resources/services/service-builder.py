from jinja2 import Environment, FileSystemLoader

file_loader = FileSystemLoader('templates')
env = Environment(loader=file_loader)
template = env.get_template('service.j2')

services = ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h']

for service in services:
    output = template.render(service=service,
                             replicas=1,
                             tag='1.0.0')
    print(output)

    filename = "service-%s%s" % (service, '.yaml')
    with open(filename, "w") as f:
        f.write(output)
