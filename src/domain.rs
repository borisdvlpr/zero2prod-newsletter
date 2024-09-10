use unicode_segmentation::UnicodeSegmentation;

pub struct NewSubscriber {
    pub email: String,
    pub name: SubscriberName,
}

pub struct SubscriberName(String);

impl SubscriberName {
    // the caller gets a shared reference to the inner string, giving it a **read-only**
    // access, disabling it to compromise our invariants.
    pub fn inner_ref(&self) -> &str {
        &self.0
    }

    // returns an instance of 'SubscriberName' if the input satisfies all
    // our validation constraints on subscriber names. It panics otherwise.
    pub fn parse(s: String) -> SubscriberName {
        let is_empty_or_whitespace = s.trim().is_empty();

        // a graphmeme is defined by the unicode standard as a "user-perceived"
        // character: 'æ' is a single graphmeme, but it is composed of two characters, 'a' and 'e'
        // and it returns an iterator over the graphmemes of the input 's'.
        // 'true' specifies that we want to use the extended graphmeme definition set,
        // the recommended one
        let is_too_long = s.graphemes(true).count() > 256;

        let forbidden_characters = ['/', '(', ')', '"', '<', '>', '\\', '{', '}'];
        let contains_forbidden_characters = s.chars().any(|g| forbidden_characters.contains(&g));

        if is_empty_or_whitespace || is_too_long || contains_forbidden_characters {
            panic!("{} is not a valid subscriber name.", s)
        
        } else {
            Self(s)
        }
    }
}
