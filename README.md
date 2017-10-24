# lpmgt
lpmgt - A Command Line Tool that manages LastPass Enterprise using LastPass Provisioning API. This CLI helps to create/read/update/delete users/members and groups under your team/company. All outputs is in JSON format, so users can pipe the outputs using number of tools such as [jq](http://stedolan.github.io/jq/).

# Setup
## Prerequisite
This tool is only available to groups who contracts LastPass Enterprise API. If you meet the condition, obtain your `companyID` and `provisioningHash` from LastPass dashboard. 

## Mac
## Windows
## Source Build
```
$ go get github.com/moneyforward/manage_lastpass
# go install github.com/moneyforward/manage_lastpass
```

# Usage
First, you need to set up couple of environment variables. Obtain those credentials from LastPass dashboard.
```
% export LASTPASS_COMPANY_ID={YOUR COMPANY ID}
% LASTPASS_APIKEY={YOUR PROVISIONING HASH from dashboard}
```

Or, you may rename config_ex.yaml as config.yaml and put replace relevant values.
```
company_id: {COMPANY_ID}
end_point_url: https://lastpass.com/enterpriseapi.php
secret: {SECRET/API_KEY}
```

## Examples
```
lpmgt get groups
lpmgt get users
lpmgt get users -f non2fa
lpmgt create user <member@email.com> -d "Department" 
lpmgt create user <member@email.com> --bulk users.json
lpmgt update user transfer <member@email.com> --leave "departmentA" --join "departmentB"
lpmgt describe user <member@email.com>
lpmgt delete user <member@email.com> --mode delete
lpmgt --config config.yaml -t ASIA/TOKYO get dashboard 
```
# Limitation
One cannot create/delete/update group info because API is not prepared in LastPass

# Contribution
1. Fork
2. Create a branch
3. Create a PR.

# License
MIT
