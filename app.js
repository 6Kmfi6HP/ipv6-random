const getRandomBinarySegment = (length) => {
    let segment = '';
    for (let i = 0; i < length; i++) {
        segment += Math.floor(Math.random() * 2);
    }
    return segment;
};

const expandIPv6Segments = (segments) => {
    const emptySegmentIndex = segments.indexOf('');
    if (emptySegmentIndex !== -1) {
        const emptySegmentCount = 8 - segments.length + 1;
        const emptySegments = Array(emptySegmentCount).fill('0000');
        return [...segments.slice(0, emptySegmentIndex), ...emptySegments, ...segments.slice(emptySegmentIndex + 1)];
    } else {
        return segments;
    }
};

const segmentToBinary = (segment) => {
    if (segment.length < 4) {
        segment = segment.padStart(4, '0');
    }
    return parseInt(segment, 16).toString(2).padStart(16, '0');
};

const convertIPv6ToBinary = (ipv6) => {
    const segments = ipv6.split(':');
    const expandedSegments = expandIPv6Segments(segments);
    const binarySegments = expandedSegments.map(segmentToBinary);
    return binarySegments.join('');
};

const padBinary = (binary) => {
    const paddedLength = Math.ceil(binary.length / 128) * 128;
    return binary.padEnd(paddedLength, '0');
};

const splitBinaryIntoSegments = (binary, segmentSize) => {
    return binary.match(new RegExp(`.{1,${segmentSize}}`, 'g'));
};

const convertBinarySegmentToHex = (segment) => {
    const hexSegment = parseInt(segment, 2).toString(16);
    return hexSegment === '0' ? '' : hexSegment;
};

const findLongestEmptySequence = (segments) => {
    let maxEmptyStart = -1;
    let maxEmptyCount = 0;
    let emptyStart = -1;
    let emptyCount = 0;

    segments.forEach((segment, index) => {
        if (segment === '') {
            if (emptyCount === 0) {
                emptyStart = index;
            }
            emptyCount++;
            if (emptyCount > maxEmptyCount) {
                maxEmptyCount = emptyCount;
                maxEmptyStart = emptyStart;
            }
        } else {
            emptyCount = 0;
        }
    });

    return { maxEmptyStart, maxEmptyCount };
};

const shortenIPv6Segments = (segments, maxEmptyStart, maxEmptyCount) => {
    if (maxEmptyCount > 1) {
        segments.splice(maxEmptyStart, maxEmptyCount, '');
    }
    return segments;
};

const convertBinaryToIPv6 = (binary) => {
    const paddedBinary = padBinary(binary);
    const segments = splitBinaryIntoSegments(paddedBinary, 16);

    const ipv6Segments = segments.map(convertBinarySegmentToHex);

    const { maxEmptyStart, maxEmptyCount } = findLongestEmptySequence(ipv6Segments);
    const shortenedSegments = shortenIPv6Segments([...ipv6Segments], maxEmptyStart, maxEmptyCount);

    return shortenedSegments.join(':');
};

const generateIPv6Commands = (input, ipv6Count, interfaceName) => {
    const [ipv6WithPrefix, prefixLength] = input.split('/');
    const fullBinary = convertIPv6ToBinary(ipv6WithPrefix);
    const prefixBinary = fullBinary.substring(0, prefixLength);

    const generatedIPv6Commands = [];

    for (let i = 0; i < ipv6Count; i++) {
        const randomBinary = getRandomBinarySegment(128 - prefixLength);
        const generatedBinary = prefixBinary + randomBinary;
        const generatedIPv6 = convertBinaryToIPv6(generatedBinary);

        generatedIPv6Commands.push(generatedIPv6 + '/' + prefixLength);
    }

    return generatedIPv6Commands;
};

const generateShellCommands = (ipv6Addresses, interfaceName) => {
    let shellCommands = '';

    for (let i = 0; i < ipv6Addresses.length; i++) {
        const ipv6Address = ipv6Addresses[i].trim();
        if (ipv6Address !== '') {
            shellCommands += `sudo ip addr add ${ipv6Address} dev ${interfaceName};`;
        }
    }

    return shellCommands;
};

const main = () => {
    if (process.argv.length < 5) {
        console.log('Usage: node script.js <input> <ipv6Count> <interfaceName>');
        process.exit(1);
    }

    const input = process.argv[2];
    const ipv6Count = parseInt(process.argv[3], 10);
    const interfaceName = process.argv[4];

    const generatedIPv6Commands = generateIPv6Commands(input, ipv6Count, interfaceName);
    const shellCommands = generateShellCommands(generatedIPv6Commands, interfaceName);

    console.log(shellCommands);
};

main();
