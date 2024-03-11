
domains=(google.com google.cn google.uk google.de google.in google.uk.co nic.ir google.ru google.org google.net google.co google.us)

for domain in ${domains[@]}; do
    echo "Testing $domain"
    whois $domain > ./test_domains/$domain.0000000000
done
