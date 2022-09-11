using AuthenticatorApi.Common.Interfaces;
using AuthenticatorApi.Data;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;

namespace AuthenticatorApi.Services.Tests;

public abstract class ServiceTestBase<T> where T : ServiceBase<T>
{
    protected IServiceCollection ServiceCollection { get; set; }
    protected IServiceProvider ServiceProvider { get; set; }

    protected Mock<IApplicationUserResolver> ApplicationUserResolverMock { get; private set; } = new Mock<IApplicationUserResolver>();
    
    public T Service { get; set; }

    public ServiceTestBase()
    {
        ServiceCollection = new ServiceCollection();
        ServiceCollection.AddAutoMapper(typeof(IAuthenticatorApiCommon));
        ServiceCollection.AddLogging();
        ServiceCollection.AddTransient<T>();
        ServiceCollection.AddDbContext<ApplicationDb>(opt => opt.UseInMemoryDatabase(Guid.NewGuid().ToString()));
        
        ServiceCollection.AddTransient<IApplicationUserResolver>(_ => ApplicationUserResolverMock.Object);
        
        Init();

        ServiceProvider = ServiceCollection.BuildServiceProvider();
        Service = ServiceProvider.GetService<T>() ?? throw new NullReferenceException($"Unable to get service {typeof(T)}");
    }
    
    /// <summary>
    /// To be implemented by any child class
    /// </summary>
    public abstract void Init();
    
    public ApplicationDb GetDbContext() => ServiceProvider.GetService<ApplicationDb>() ??
                                           throw new NullReferenceException("Could not get DB from service provider");
}