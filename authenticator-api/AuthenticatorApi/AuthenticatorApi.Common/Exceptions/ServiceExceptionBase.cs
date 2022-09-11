using System.Runtime.Serialization;
using AuthenticatorApi.Common.Enums;

namespace AuthenticatorApi.Common.Exceptions;

public abstract class ServiceExceptionBase : Exception
{
    public abstract int ResponseCode { get; }
    public abstract string DefaultMessage { get; }
    public override string Message => base.Message ?? DefaultMessage;
    public ErrorCode? ErrorCode { get; set; }
    
    protected ServiceExceptionBase() { }

    protected ServiceExceptionBase(SerializationInfo info, StreamingContext context) : base(info, context) { }

    protected ServiceExceptionBase(string message) : base(message) { }

    protected ServiceExceptionBase(string message, Exception innerException) : base(message, innerException) { }

    protected ServiceExceptionBase(ErrorCode errorCode) : base(errorCode.ToString())
    {
        ErrorCode = errorCode;
    }
}