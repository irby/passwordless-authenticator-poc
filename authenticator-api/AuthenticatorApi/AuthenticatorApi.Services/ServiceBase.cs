namespace AuthenticatorApi.Services;

public abstract class ServiceBase<T> where T : class
{
    protected ILogger<T> Log { get; set; }
    protected IMapper Mapper { get; set; }
    
    protected ServiceBase(ILoggerFactory loggerFactory, IMapper mapper)
    {
        Log = loggerFactory.CreateLogger<T>();
        Mapper = mapper;
    }
}