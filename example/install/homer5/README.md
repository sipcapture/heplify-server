### Manual installation of heplify-server + homer5 ui

#### Tested with the smallest debian 9.4 x64 droplet on digital ocean 
* login and change your password
* run inside the droplet following commands (best one by one) on console 

```
apt update && apt -y upgrade
apt -y install vim dirmngr git wget php7.0-fpm php7.0-mysql php7.0-xml nginx tcpdump mariadb-server

mysql -u root
update mysql.user set plugin=' ' where User='root';flush privileges;exit;

wget https://github.com/sipcapture/heplify-server/releases/download/0.98/heplify-server
chmod +x heplify-server
./heplify-server &

git clone https://github.com/sipcapture/homer-ui.git
cp -r homer-ui/* /var/www/html/
git clone https://github.com/sipcapture/homer-api.git
cp -r homer-api/* /var/www/html/

mv /var/www/html/api/preferences_example.php /var/www/html/api/preferences.php
mv /var/www/html/api/configuration_example.php /var/www/html/api/configuration.php

chown -R www-data:www-data /var/www/html/store/
chmod -R 0775 /var/www/html/store/dashboard
```

* remove everything inside /etc/nginx/sites-available/default 
* paste into the empty /etc/nginx/sites-available/default 
```
server {
    listen 80 default_server;
    server_name _;

    root /var/www/html;
    index index.html index.htm;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        include fastcgi_params;
        fastcgi_index index.php;
        fastcgi_read_timeout 30;
        fastcgi_pass unix:/run/php/php7.0-fpm.sock;
        fastcgi_intercept_errors on;
        fastcgi_param SCRIPT_FILENAME $document_root/api/index.php;
    }
}
```

* save and run
```
/etc/init.d/nginx restart
```

* remove everything inside /etc/php/7.0/fpm/pool.d/www.conf 
* paste into the empty /etc/php/7.0/fpm/pool.d/www.conf 
```
[www]
user = www-data
group = www-data
listen = /run/php/php7.0-fpm.sock
listen.owner = www-data
listen.group = www-data
listen.allowed_clients = 127.0.0.1
pm = static
pm.max_children = 40
pm.start_servers = 15
pm.min_spare_servers = 15
pm.max_spare_servers = 35
slowlog = log/$pool.log.slow
request_terminate_timeout = 30
chdir = /
catch_workers_output = yes
php_admin_value[error_log] = /var/log/fpm-php.www.log
php_admin_flag[log_errors] = on
```

* save and run
```
/etc/init.d/php7.0-fpm restart
```

* point your browser to your dropletIP and login with admin/test123
* use heplify to send traffic to heplify-server
* ./heplify -i any -hs dropletIP:9060 -nt tls
* change the time inside the homer ui to last 30 minutes (in the top right corner)
* have fun