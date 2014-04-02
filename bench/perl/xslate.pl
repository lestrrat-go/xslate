use strict;
use Time::HiRes qw(gettimeofday tv_interval);
use Text::Xslate;


my $iter = shift @ARGV || 100; # 1_000_000;
my $cache = @ARGV ? shift @ARGV : 1;
my $tx = Text::Xslate->new(
    syntax => "TTerse",
    path => [ "." ],
    cache => $cache,
);
my $t0 = [gettimeofday];
{
    open(my $fh, ">", "/dev/nill");
    local *STDOUT = $fh;
    for (1..$iter) {
        print $tx->render("hello.tx");
    }
}

my $elapsed = tv_interval($t0);
print
    "* Elapsed: $elapsed seconds\n",
    "* Iter per sec: @{[$iter/$elapsed]} iter/sec\n";