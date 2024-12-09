#include <iostream>
#include <sstream>
#include <vector>
#include <string>
#include <krb5.h>

void Cleanup(_krb5_context* context,
                    krb5_principal_data* server,
                    krb5_principal_data* client,
                    _krb5_ccache* ccache,
                    _krb5_auth_context* auth_context,
                    _krb5_creds* credsp,
                    char* mrealm) {

    if(context == nullptr) return;
    if(client != nullptr) krb5_free_principal(context, client);
    if(ccache != nullptr) krb5_cc_close(context, ccache);
    if(credsp != nullptr) krb5_free_creds(context, credsp);
    if(mrealm != nullptr) krb5_free_default_realm(context, mrealm);
    krb5_free_context(context);
}
int Authenticate(std::string& host, const std::string& service_name,
                 std::vector<unsigned char> *outBuffer, std::string* errorMsg) {
    _krb5_context* context = nullptr;
    krb5_principal_data* server = nullptr;
    char* mrealm = nullptr;
    _krb5_ccache* ccache = nullptr;
    krb5_principal_data* client = nullptr;
    _krb5_auth_context* auth_context = nullptr;
    _krb5_creds* credsp = nullptr;
    unsigned char* output = nullptr;
    std::stringstream  ss;
    _krb5_data krb5Data2{};
    _krb5_creds krb5Creds{};
    _krb5_data krb5Data1{0, static_cast<unsigned int>(host.length()), host.data()};
    int result = krb5_init_context(&context);
    if(result != 0 ) {
        ss << "error in krb5_init_context: " << result;
        goto clean;
    }
    result = krb5_sname_to_principal(context, host.c_str(), service_name.c_str(), 3, &server);
    if (result != 0) {
        ss << "error in krb5_sname_to_principal: " << krb5_get_error_message(context, result);;
        goto clean;
    }
    // load default cache is enough
    result = krb5_cc_default(context, &ccache);
    if (result != 0) {
        ss << "error in krb5_cc_resolve(): " << krb5_get_error_message(context, result);;
        goto clean;
    }

    result = krb5_cc_get_principal(context, ccache, &client);
    if (result != 0) {
        ss << "error in krb5_cc_get_principal: " << krb5_get_error_message(context, result);;
        goto clean;
    }
    result = krb5_auth_con_init(context, &auth_context);
    if (result != 0) {
        ss << "error in krb5_auth_con_init: " << krb5_get_error_message(context, result);
        goto clean;
    }
    memset(&krb5Creds, 0, sizeof(_krb5_creds));
    result = krb5_copy_principal(context, server, &krb5Creds.server);
    if (result != 0) {
        ss << "error in krb5_copy_principal: " << krb5_get_error_message(context, result);;
        goto clean;
    }
    result = krb5_copy_principal(context, client, &krb5Creds.client);
    if (result != 0) {
        ss << "error in krb5_copy_principal: " << krb5_get_error_message(context, result);;
        goto clean;
    }
    result = krb5_get_credentials(context, 0, ccache, &krb5Creds, &credsp);
    if (result != 0) {
        ss << "error in krb5_get_credentials: " << krb5_get_error_message(context, result);
        goto clean;
    }

    result = krb5_mk_req_extended(context, &auth_context, 0x20000000, &krb5Data1, credsp, &krb5Data2);
    if (result != 0) {
        ss << "error in krb5_mk_req_extended: " << krb5_get_error_message(context, result);
        goto clean;
    }
    outBuffer->resize(krb5Data2.length);
    memcpy(outBuffer->data(), krb5Data2.data, krb5Data2.length);
    clean:
    Cleanup(context, server, client, ccache, auth_context, credsp, mrealm);
    if(ss.str().length() > 0) {
        *errorMsg = ss.str();
        return 1;
    }
    return 0;
}