import socket
import struct

import netaddr
from socip_build import SocDBEncoder


def build_cidr(minip, maxip):
    minip, maxip = tuple(map(lambda x: socket.inet_ntoa(
        struct.pack('!L', int(x))), (minip, maxip)))
    # print(minip, maxip)
    for c in netaddr.iprange_to_cidrs(minip, maxip):
        yield str(c)


def load_data(filename):
    fields = [
        'minip', 'maxip', 'continent', 'areacode', 'adcode', 'country',
        'province', 'city', 'district', 'bd_lon', 'bd_lat', 'wgs_lon',
        'wgs_lat', 'radius', 'scene', 'accuracy', 'owner'
    ]
    with open(filename) as fp:
        for line in fp:
            yield {f: v for f, v in zip(fields, line.replace('\n', '').split('\t'))}


if __name__ == '__main__':
    source_file = "ip_source.txt"
    ip_version = 4
    pointer_size = 32
    db_name = 'SOC-GeoIP'
    languages = ['en']
    description = {'en': 'Soc custom GeoIP databases'}
    encoder = SocDBEncoder(ip_version, pointer_size,
                           db_name, languages, description, compat=True)
    for item in load_data(source_file):
        for cidr in build_cidr(item['minip'], item['maxip']):
            build_data = encoder.insert_data(item)
            encoder.insert_network(cidr, build_data)
            encoder.write_file('py_build.socdb')
