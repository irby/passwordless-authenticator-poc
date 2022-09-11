using System.Net;

namespace AuthenticatorApi.Common.Exceptions;

public class NotFoundException: ServiceExceptionBase
{
    public override int ResponseCode => (int)HttpStatusCode.NotFound;
    public override string DefaultMessage => "";
}