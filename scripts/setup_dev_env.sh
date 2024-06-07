ip l add br-test type bridge
ip a add 10.66.0.254/24 dev br-test
ip l set br-test up

ip netns add c1
ip netns exec c1 ip l set lo up
ip l add hTc1 type veth peer name c1Th
ip l set hTc1 master br-test
ip l set hTc1 up
ip l set c1Th netns c1
ip netns exec c1 ip a add 10.66.0.1/24 dev c1Th
ip netns exec c1 ip l set c1Th up

ip netns add c2
ip netns exec c2 ip l set lo up
ip l add hTc2 type veth peer name c2Th
ip l set hTc2 master br-test
ip l set hTc2 up
ip l set c2Th netns c2
ip netns exec c2 ip a add 10.66.0.2/24 dev c2Th
ip netns exec c2 ip l set c2Th up