use strict;
use Time::HiRes qw(gettimeofday tv_interval);
use Text::Xslate;
use Getopt::Long;

my $iter;
my $cache;

if (! GetOptions(
    "iterations=i" => \$iter,
    "cacheLevel=i" => \$cache,
)) {
    exit 1;
}

my $tx = Text::Xslate->new(
    syntax => "TTerse",
    path => [ "." ],
    cache => $cache,
);
my $t0 = [gettimeofday];
{
    open(my $fh, ">", "/dev/null") or die "$!";
    local *STDOUT = $fh;
    for (1..$iter) {
        print $tx->render("hello.tx");
    }
}

my $elapsed = tv_interval($t0);
printf "* Elapsed: %f seconds\n", $elapsed;
printf "* Secs per iter: %f secs/iter\n", @{[$elapsed/$iter]};
printf "* Iter per sec: %f iter/sec\n", @{[$iter/$elapsed]};