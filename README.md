# go-ioc

A reflection-based library with the ability to generate, register, and manage the retrieval of beans by type from a context, similar to a service locator.

## Usage:

Optionally, set up a context in your project as follows:

```go
    package appPackage
    import 	. "github.com/catmorte/go-inversion_of_control/pkg/context"

    func init() {
      SetContext(NewMemoryContext())
    }
```

This creates an instance of the default context implementation from the package:

    github.com/catmorte/go-inversion_of_control/pkg/context

The next step is to define your beans in the project:

```go
    package appPackage/beans
    import (
      ...
      . "github.com/catmorte/go-inversion_of_control/pkg/context"
    )

    func init() {
      // start dependencies definition
      dep1 := Dep[*dep1Type]()
      dep2 := Dep[*dep1Type]()
      dep3 := DepInterface[interfaceType1]()
      ...
      depN := Dep[*depNType]()
      // end dependencies definition


      // start bean constructor definition
      Reg[*beanType](func() *beanType {
        return NewBeanType(
                ResolveDep[*dep1Type](dep1),
                ResolveDep[*dep2Type](dep2),
                ResolveDep[interfaceType1](dep3),
                ...
                ResolveDep[*depNType](depN),
        			)
      }, dep1, dep2, dep3, ..., depN)
      // end bean constructor definition
    }

```

The **Reg** function initializes the bean and waits for all the required dependencies to be available.

> [!NOTE]  
> Ensure that you pass all dependencies to the Reg function and use them in the constructor, or bean initialization will be blocked.

Next, import the context and beans initialization:

```go
    import (
        ...
         _ "appPackage"        // initialize the context
         _ "appPackage/beans"  // initialize beans
        ...
     )
```

> [!NOTE]  
> Ensure that the context initialization is imported before the beans import.

To retrieve a bean, use the following in your project:

```go
bean := Ask[*beanType]()
```

The **Ask** function waits until the bean is initialized, then retrieves it.

```go
bean := AskInterface[intrfaceType]()
```

**AskInterface** is used when you need an interface implementation (duck typing). Ensure you ask for as specific an interface as possible to avoid unexpected results.

> [!CAUTION]  
> Make sure to use the interface type in the type argument; otherwise, it will cause a panic.

You can find a working example in the [/example folder](https://github.com/catmorte/go-ioc/tree/main/examples).

The library also supports named scopes:
**RegScoped**, **AskScoped**, **DepScoped**, **AskInterfaceScoped**, **DepInterfaceScoped**

---

## Code Generation

It's possible to generate boilerplate code based on structure definitions. To do this, follow these steps:

- Install `go-ioc`
- Add `//go:generate go-ioc` to the file
- Add {{strategy}}.Bean[resultType] to the structure for which code generation is needed, where **{{strategy}}** is either singleton from `github.com/go-ioc/pkg/context/singleton` or prototype from `github.com/go-ioc/pkg/context/prototype`.
- Add the `bean:""` tag to the root fields that require injection. (To use a non-default scope, specify the scope name in the tag like `bean:"someScope"`). If the tag is defined for the Bean itself, the bean will be scoped accordingly. You can also define interface injections by setting interface as the second value in the tag, e.g., `bean:"someScope,interface"` or `bean:",interface"`.
- call `go generate ./...`
- Finally, import all the necessary packages in your main.go like so:

```go
import _ "some/package/with/bean/declaration"
```

You can find a working example in the [/example folder](https://github.com/catmorte/go-ioc/tree/main/examples).
