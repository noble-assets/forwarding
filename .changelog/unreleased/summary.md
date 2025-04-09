*Apr 9, 2025*

This is a non-consensus breaking patch release to the v2 line. The release
improve the constructor of the `SigVerificationDecorator`
to accept an interface for the underlying ante handler. This change allows
to chain multiple signature verification decorator for
other modules.
