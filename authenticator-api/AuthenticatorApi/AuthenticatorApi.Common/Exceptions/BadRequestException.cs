using System.Net;
using AuthenticatorApi.Common.Enums;

namespace AuthenticatorApi.Common.Exceptions;

public class BadRequestException: ServiceExceptionBase
{
    public override int ResponseCode => (int)HttpStatusCode.BadRequest;
    public override string DefaultMessage => base.Message;
    public BadRequestException() {}
    public BadRequestException(ErrorCode code) : base(code) {}
}