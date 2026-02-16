# How to use
### define env vars:
* `USER`=username
* `PASSWORD`=your_password
* `SERVER`=ip or servername
* `PORT`=default 1521
* `SERVICE`=service_name
* `SSL`=TRUE if your connection is secure otherwise false
* `WALLET`=path to wallet should be present if SSL=TRUE

### test all issue
run test for entire folder
### test only some issues or feature
* create new folder 
* copy `global_var.go`
* copy required test files
* run test on the folder