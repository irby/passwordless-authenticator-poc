using AuthenticatorApi.Api.Handlers;
using AuthenticatorApi.Common.Constants;
using AuthenticatorApi.Common.Interfaces;
using AuthenticatorApi.Data;
using AuthenticatorApi.Data.Seed;
using AuthenticatorApi.Services.Domain;
using Microsoft.EntityFrameworkCore;
using Moq;

var builder = WebApplication.CreateBuilder(args);

// Add services to the container.

builder.Services.AddControllers();
// Learn more about configuring Swagger/OpenAPI at https://aka.ms/aspnetcore/swashbuckle
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();

ConfigureServices(builder.Services);
RunMigrations(builder.Services);

var app = builder.Build();

ConfigureMiddleware(app);

// Configure the HTTP request pipeline.
if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI();
}

app.UseHttpsRedirection();

app.UseAuthorization();

app.MapControllers();

app.Run();



void ConfigureServices(IServiceCollection collection)
{
    collection.AddTransient<IApplicationUserResolver>(_ =>
    {
        var resolver = new Mock<IApplicationUserResolver>();
        return resolver.Object;
    });
    collection.AddDbContext<ApplicationDb>(opt => opt.UseInMemoryDatabase(StringConstants.AuthenticationDbName));
    collection.AddTransient<IUserService, UserService>();
    collection.AddTransient<AuthenticationService>();
    collection.AddAutoMapper(typeof(IAuthenticatorApiCommon).Assembly);
}

void ConfigureMiddleware(IApplicationBuilder builder)
{
    builder.UseServiceExceptionHandler();
}

void RunMigrations(IServiceCollection collection)
{
    var provider = collection.BuildServiceProvider();
    var db = provider.GetService<ApplicationDb>();
    db.Seed();
}