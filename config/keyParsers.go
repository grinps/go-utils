package config

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const DefaultKeyNameValidationRegEx = "^[a-zA-Z0-9]*$"
const DefaultResolveFormatRegEx = "^\\$\\{\\{([a-zA-Z0-9]*)\\}\\}$"

var keyNameValidRegEx = regexp.MustCompile(DefaultKeyNameValidationRegEx)
var resolveFormatRegEx = regexp.MustCompile(DefaultResolveFormatRegEx)

type OptionKeyParser func(ctx context.Context, parserOption *keyParserOptions) error

type keyParserOptions struct {
	keyStringRegex *regexp.Regexp
	delimiter      string
	resolverRegex  *regexp.Regexp
	resolvers      []KeyNameResolver
}

func DefaultKeyParser(ctx context.Context, options ...OptionKeyParser) KeyParser {
	parserOption := &keyParserOptions{
		keyStringRegex: keyNameValidRegEx,
		delimiter:      DefaultKeyDelimiter,
		resolverRegex:  resolveFormatRegEx,
		resolvers:      []KeyNameResolver{},
	}
	var parseSetupErrs []error
	for _, option := range options {
		setupErr := option(ctx, parserOption)
		if setupErr != nil {
			parseSetupErrs = append(parseSetupErrs, setupErr)
		}
	}
	var parserErrs error
	if len(parseSetupErrs) > 0 {
		parserErrs = errors.Join(parseSetupErrs...)
	}
	return func(ctx context.Context, keyParsers []KeyParser, key string, previousParsedKeys []Key) (parsedKey []Key, parsingErr error) {
		const parseErrString = "failed to parse key %s due to error %w"
		if parserErrs != nil {
			return []Key{}, fmt.Errorf(parseErrString, key, parserErrs)
		}
		if key == "" {
			return []Key{}, fmt.Errorf(parseErrString, key, errors.New("given key name is an empty string"))
		}
		keyItems := strings.Split(key, parserOption.delimiter)
		if len(keyItems) == 0 {
			return []Key{}, fmt.Errorf(parseErrString, key, fmt.Errorf("no key items generated from splitting"))
		}
		var generationErrors []error
		var generatedKeyItems []Key
		for _, keyItem := range keyItems {
			if keyItem != "" {
				if parserOption.keyStringRegex != nil && !parserOption.keyStringRegex.MatchString(keyItem) {
					generationErrors = append(generationErrors,
						fmt.Errorf("key element %s does not matches regular expression %s", keyItem, parserOption.keyStringRegex.String()))
				} else if len(parserOption.resolvers) > 0 && parserOption.resolverRegex != nil && parserOption.resolverRegex.MatchString(keyItem) {
					extractedItems := parserOption.resolverRegex.FindStringSubmatch(keyItem)
					if len(extractedItems) == 2 {
						var resolveKeyDetail = KeyResolutionDetail{}
						resolveKeyDetail[0] = extractedItems[0]
						resolveKeyDetail[1] = extractedItems[1]
						var resolvedKeys []Key
						for resolverIndex, resolver := range parserOption.resolvers {
							keys, resolutionErr := resolver(ctx, parserOption.resolvers, resolveKeyDetail, resolvedKeys)
							if resolutionErr != nil && len(keys) > 0 {
								// Adding the generated keys as additional key values and stopping resolution process
								// TODO: there are alternate implementations possible that we are avoiding here.
								generatedKeyItems = append(generatedKeyItems, keys...)
								break
							} else {
								generationErrors = append(generationErrors, fmt.Errorf("failed to resolve key element %s with resolver number %d due to error %w", keyItem, resolverIndex, resolutionErr))
							}
						}
					} else {
						generationErrors = append(generationErrors,
							fmt.Errorf("expected 2 items from matching key item %s with resolver regex %s, actual %d", keyItem, parserOption.resolverRegex.String(), len(extractedItems)))
					}
				} else {
					generatedKeyItems = append(generatedKeyItems, simpleStringKey(keyItem))
				}
			} else {
				generationErrors = append(generationErrors, fmt.Errorf("empty key element found in key name %s", key))
			}
		}
		if generationErrors != nil {
			return []Key{}, fmt.Errorf(parseErrString, key, errors.Join(generationErrors...))
		}
		generatedKeyItemLength := len(generatedKeyItems)
		var generatedKey Key
		switch {
		case generatedKeyItemLength == 1:
			generatedKey = generatedKeyItems[0]
		case generatedKeyItemLength > 1:
			var asComplexKey = ComplexKey(generatedKeyItems)
			generatedKey = &asComplexKey
		default:
			generatedKey = InvalidKey
		}
		parsedKey = append(parsedKey, generatedKey)
		return
	}
}

func OptionKeyParserKeyNameRegEx(regExp string) OptionKeyParser {
	return func(ctx context.Context, parserOption *keyParserOptions) error {
		compiledRegEx, compileErr := regexp.Compile(regExp)
		if compileErr == nil {
			parserOption.keyStringRegex = compiledRegEx
		}
		return compileErr
	}
}

func OptionKeyParserHierarchyKeyDelimiter(delimiter string) OptionKeyParser {
	return func(ctx context.Context, parserOption *keyParserOptions) error {
		if delimiter == "" {
			return errors.New("delimiter for detecting hierarchical key was empty")
		}
		if len(delimiter) != 1 {
			return fmt.Errorf("delimiter for detecting hierarchical key was %s which is multi-character", delimiter)
		}
		parserOption.delimiter = delimiter
		return nil
	}
}

func OptionKeyParserWithKeyResolvers(resolvers ...KeyNameResolver) OptionKeyParser {
	return func(ctx context.Context, parserOption *keyParserOptions) error {
		parserOption.resolvers = resolvers
		return nil
	}
}

func OptionKeyParserResolveFormat(regExp string) OptionKeyParser {
	return func(ctx context.Context, parserOption *keyParserOptions) error {
		compiledRegEx, compileErr := regexp.Compile(regExp)
		if compileErr == nil {
			parserOption.resolverRegex = compiledRegEx
		}
		return compileErr
	}
}
