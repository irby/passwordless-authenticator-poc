using System.Net;

namespace AuthenticatorApi.Common.Exceptions;

public class NotAuthenticatedException : ServiceExceptionBase
{
    public override int ResponseCode => (int)HttpStatusCode.Unauthorized;
    public override string DefaultMessage => "You are not authenticated. Please login and try your request again";
}