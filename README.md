# go-ioc

Yet another library based on reflection with the ability to generate and manage the registration and retrieval of beans by type from a context, similar to a service locator.

### How to use:

Optional set context by doing the following in your project:

    package appPackage
    import 	. "github.com/catmorte/go-inversion_of_control/pkg/context"

    func init() {
      SetContext(NewMemoryContext())
    }

It creates an instance of default context implementation from package:

    github.com/catmorte/go-inversion_of_control/pkg/context

The next step is to create beans by doing the following in your project:

    package appPackage/beans
    import (
      ...
      . "github.com/catmorte/go-inversion_of_control/pkg/context"
    )

    func init() {
      // start dependencies definition
      dep1 := Dep[*dep1Type]()
      dep2 := Dep[*dep1Type]()
      ...
      depN := Dep[*depNType]()
      // end dependencies definition


      // start bean constructor definition
      Reg[*beanType](func() *beanType {
        return NewBeanType(
                ResolveDep[*dep1Type](dep1),
                ResolveDep[*dep2Type](dep2),
                ...
                ResolveDep[*depNType](depN),
        			)
      }, dep1, dep2, ..., depN)
      // end bean constructor definition
    }

**Reg** function starts bean initialization and waits until all the necessary dependencies will be initialized.

**! Please ensure that you pass all the deps in the function Reg as well as you using them inside constructor, otherwise it will block bean initialization**

The next step is to import context and beans initialization:

    import (
        ...
         _ "appPackage"        // initialize the context
         _ "appPackage/beans"  // initialize beans
        ...
     )

**! Make sure that the context initialization imported before beans import**

To retrieve bean use the following in your project:

`bean := Ask[*beanType]()`

**Ask** function waits until bean will be initialized and then retrieve it

You may find the working example in folder /example

Also supported **named scopes**: **RegScoped, AskScoped, DepScoped**


---

It's possible to generate boilerplate code based on the structure definition. To make it possible it's necessary to do 6 things:
- install `go-ioc`
- add `//go:generate go-ioc` to the file
- add `{{strategy}}.Bean[resultType]` to the structure for which code-generation needed where `{{strategy}}` is either `singleton` from `github.com/go-ioc/pkg/context/singleton` or `prototype` from `github.com/go-ioc/pkg/context/prototype`. 
- add tag `bean:""` to the root fields which require injection. (to use non-nil scope put scope name into the tag like so `bean:"some scope"`). If tag defined for `Bean` itself - then bean will be defined in this scope rather then default one.
- call `go generate ./...`
- in the end - import all the necessary packages in the `main.go` like so `import _ "some/package/with/bean/declaration"`

check out `examples` folder
