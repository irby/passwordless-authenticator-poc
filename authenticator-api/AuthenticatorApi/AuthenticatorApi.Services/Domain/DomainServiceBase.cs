using AuthenticatorApi.Data;

namespace AuthenticatorApi.Services.Domain;

public abstract class DomainServiceBase<T> : ServiceBase<T> where T : class
{
    protected ApplicationDb Db { get; set; }

    protected DomainServiceBase(ILoggerFactory loggerFactory, ApplicationDb db, IMapper mapper) : base(loggerFactory, mapper)
    {
        Db = db;
    }
}