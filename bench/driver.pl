use strict;

my @perl_cmd = qw(perl perl/xslate.pl);
my @go_cmd   = qw(./go/xslate);
my %cmds     = (
    "p5-xslate" => \@perl_cmd,
    "go-xslate" => \@go_cmd,
);

if (! -f "./go/xslate") {
    chdir "go";
    system "go", "build", "xslate.go";
    chdir "..";
}

foreach my $cache (0, 1) {
    for my $lang ( "p5-xslate", "go-xslate" ) {
        print "# $lang (cache @{[ $cache ? 'ENABLED' : 'DISABLED' ]})\n";
        for my $iter (10, 100, 1000, 10000) {
            print "iter ($iter)\n";
            system @{$cmds{$lang}}, $iter, $cache;
        }
        print "====\n";
    }
}