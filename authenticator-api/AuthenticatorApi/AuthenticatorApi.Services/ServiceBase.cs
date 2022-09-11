namespace AuthenticatorApi.Services;

public abstract class ServiceBase<T> where T : class
{
    protected ILogger<T> Log { get; set; }
    
    protected ServiceBase(ILoggerFactory loggerFactory)
    {
        Log = loggerFactory.CreateLogger<T>();
    }
}