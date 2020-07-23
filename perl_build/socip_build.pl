use MaxMind::DB::Writer::Tree;
use Net::Works::Network;

use strict;
use warnings;
use Path::Class;
use autodie;
use Net::CIDR;

# Field type definitions
my % types = (
    accuracy  => 'utf8_string',
    continent => 'utf8_string',
    country   => 'utf8_string',
    owner     => 'utf8_string',
    scene     => 'utf8_string',
    source    => 'utf8_string',
    zipcode   => 'utf8_string',
    city      => 'utf8_string',
    district  => 'utf8_string',
    lat       => 'utf8_string',
    lng       => 'utf8_string',
    province  => 'utf8_string',
    radius    => 'utf8_string',
);


# Integer into a IPv4
sub dec2addr {
    my $decimal = $_[0];
    my @bytes = unpack 'CCCC', pack 'N', $decimal;
    my $ipv4 = join '.', @bytes;
    return $ipv4;
}

# tree information definitions
my $tree = MaxMind::DB::Writer::Tree->new(
    ip_version               => 4,
    record_size              => 28,
    database_type            => 'SOC-GeoIP',
    languages                => [ 'en', 'cn' ],
    description              => {
        en => 'SOC IP geographic information offline database',
        cn => '赛欧思IP地理信息离线数据库',
    },
    remove_reserved_networks => 0, # 支持保留IP的查询
    map_key_type_callback    => sub {
        $types{
            $_[0]
        }
    },
);

# insert networks to tree
sub tree_insert_network {
    my $network = Net::Works::Network->new_from_string(string => $_[1]);
    $_[0]->insert_network(
        $network, {
        accuracy  => $_[16],
        continent => $_[3],
        country   => $_[6],
        city      => $_[8],
        district  => $_[9],
        lat       => $_[13],
        lng       => $_[12],
        province  => $_[7],
        radius    => $_[14],
        owner     => $_[17],
        scene     => $_[15],
        source    => 'soc-geoip',
        zipcode   => '',
    }
    );
}

# Main logic
sub build_tree {
    my $file = file($_[0]);
    my $file_handle = $file->openr();
    binmode($file_handle, ":utf8");
    while (my $line = $file_handle->getline()) {
        $line =~ s/^\s+|\s+$//g;
        if ($line eq "") {
            next;
        }
        $line =~ s/"//g;
        $line =~ s/;//g;
        my @values = split('\t', $line);
        if ($values[0] eq "id") {
            next
        }

        # int -> ipv4
        my $min_ip_addr = dec2addr($values[1]);
        my $max_ip_addr = dec2addr($values[2]);
        print "write -> $min_ip_addr - $max_ip_addr\n";
        # ipv4 range -> cidr
        foreach my $nets (Net::CIDR::range2cidr(join("-", $min_ip_addr, $max_ip_addr))) {
            for (my $i = 3; $i <= 18; $i++) {
                if (!defined $values[$i]) {
                    $values[$i] = "";
                }

                # load ip description information
                my @networks = split(',', $nets);
                foreach my $net (@networks) {
                    # insert data to tree
                    tree_insert_network($tree, $net,
                        $values[2], $values[3], $values[4], $values[5],
                        $values[6], $values[7], $values[8], $values[9],
                        $values[10], $values[11], $values[12], $values[13],
                        $values[14], $values[15], $values[16], $values[17]
                    );
                }
            }
        }
    }
}



my $filename = $ARGV[0] or die "Need to get CSV(,) file on the command line\n";
my $dbname = $ARGV[1] or die "Output database file needs to be specified\n";

build_tree($filename, $tree);
open my $fh, '>:bytes', $dbname;
$tree->write_tree($fh);
