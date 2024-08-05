ip l add br-test type bridge
ip a add 10.67.0.254/24 dev br-test
ip l set br-test up

ip netns add c1
ip netns exec c1 ip l set lo up
ip l add hTc1 type veth peer name c1Th
ip l set hTc1 master br-test
ip l set hTc1 up
ip l set c1Th netns c1
ip netns exec c1 ip a add 10.67.0.1/24 dev c1Th
ip netns exec c1 ip l set c1Th up

ip netns add c2
ip netns exec c2 ip l set lo up
ip l add hTc2 type veth peer name c2Th
ip l set hTc2 master br-test
ip l set hTc2 up
ip l set c2Th netns c2
ip netns exec c2 ip a add 10.67.0.2/24 dev c2Th
ip netns exec c2 ip l set c2Th up

ip netns add n1
ip netns exec n1 ip l set lo up
ip l add n1Tc2 type veth peer name c2Tn1
ip l set c2Tn1 netns c2
ip l set n1Tc2 netns n1
ip netns add n2
ip netns exec n2 ip l set lo up
ip l add n2Tc2 type veth peer name c2Tn2
ip l set c2Tn2 netns c2
ip l set n2Tc2 netns n2

ip netns exec c2 ip l add br0 type bridge
ip netns exec c2 ip l set c2Tn1 master br0
ip netns exec c2 ip l set c2Tn2 master br0
ip netns exec c2 ip a add 10.68.0.254/24 dev br0
ip netns exec c2 ip l set br0 up
ip netns exec c2 ip l set c2Tn1 up
ip netns exec c2 ip l set c2Tn2 up

ip netns exec n1 ip a add 10.68.0.1/24 dev n1Tc2
ip netns exec n1 ip l set n1Tc2 up
ip netns exec n2 ip a add 10.68.0.2/24 dev n2Tc2
ip netns exec n2 ip l set n2Tc2 up

ip netns exec c2 iptables -t nat -A POSTROUTING -o c2Th -j MASQUERADE
ip netns exec n1 ip r add default via 10.68.0.254
ip netns exec n2 ip r add default via 10.68.0.254